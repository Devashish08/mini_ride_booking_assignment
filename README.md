## Mini Ride Booking (Go + Kafka/Redpanda + Postgres)

### Stack
- Services: `booking_svc` (REST + DB + producer + consumer), `driver_svc` (REST + DB + consumer + producer)
- MQ: Redpanda (Kafka API). Topics: `booking.created`, `booking.accepted`
- DB: PostgreSQL (one DB per service)
- Go: 1.24.x
- Arch: Clean Architecture

### Run
```bash
# from repo root
cd deploy
docker compose up --build -d
# health
curl -sf localhost:8080/healthz && echo booking_svc OK
curl -sf localhost:8081/healthz && echo driver_svc OK
```

### Env vars
- booking_svc
  - `HTTP_PORT=8080`, `LOG_LEVEL=info`
  - `DB_HOST=booking_db`, `DB_PORT=5432`, `DB_USER=booking`, `DB_PASSWORD=booking`, `DB_NAME=booking`
  - `KAFKA_BROKERS=redpanda:9092`
  - `TOPIC_BOOKING_CREATED=booking.created`
  - `TOPIC_BOOKING_ACCEPTED=booking.accepted`
  - `CONSUMER_GROUP_ACCEPTS=booking_svc.accepts`
- driver_svc
  - `HTTP_PORT=8081`, `LOG_LEVEL=info`
  - `DB_HOST=driver_db`, `DB_PORT=5432`, `DB_USER=driver`, `DB_PASSWORD=driver`, `DB_NAME=driver`
  - `KAFKA_BROKERS=redpanda:9092`
  - `TOPIC_BOOKING_CREATED=booking.created`
  - `TOPIC_BOOKING_ACCEPTED=booking.accepted`
  - `CONSUMER_GROUP_JOBS=driver_svc.jobs`

### Sample curl
```bash
# create booking
curl -X POST localhost:8080/bookings \
 -H "Content-Type: application/json" \
 -d '{"pickuploc":{"lat":12.9,"lng":77.6},"dropoff":{"lat":12.95,"lng":77.64},"price":220}'

# list bookings
curl localhost:8080/bookings

# driver side
curl localhost:8081/drivers
curl localhost:8081/jobs

# accept (first wins; others 409)
curl -X POST localhost:8081/jobs/<booking_id>/accept \
 -H "Content-Type: application/json" \
 -d '{"driver_id":"d-1"}'
```

### See messages
- Redpanda Console: http://localhost:8082
- CLI:
```bash
docker compose exec redpanda rpk topic list | cat
docker compose exec redpanda rpk topic consume booking.created -n 5 -o newest | cat
docker compose exec redpanda rpk topic consume booking.accepted -n 5 -o newest | cat
```

### Tests
```bash
cd booking_svc && go test ./...
cd ../driver_svc && go test ./...
```

### Assumptions
- At-least-once processing; handlers are idempotent (`ON CONFLICT` or `WHERE status=...`).
- Ordering is per booking by using `booking_id` as Kafka key.
- Only two ride statuses are modeled: Requested, Accepted.

### Troubleshooting
- If POST /bookings returns 500 and no events, ensure topics exist and Redpanda advertises `PLAINTEXT://redpanda:9092` to in-network clients (compose already configured).
- If services start before DB is ready, compose uses `depends_on: service_healthy` and `restart: on-failure` to recover.