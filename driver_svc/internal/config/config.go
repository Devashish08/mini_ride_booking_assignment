package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName     string
	HTTPPort        string
	GracefulTimeout time.Duration
	LogLevel        string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	KafkaBrokers         string
	TopicBookingCreated  string
	TopicBookingAccepted string
	ConsumerGroupJobs    string
}

func LoadFromEnv(serviceName, defaultPort string) Config {
	logLevel := getEnv("LOG_LEVEL", "info")
	port := getEnv("HTTP_PORT", defaultPort)
	gt := getEnvInt("GRACEFUL_TIMEOUT_SECONDS", 10)

	dbHost := getEnv("DB_HOST", defByService(serviceName, "booking_db", "driver_db"))
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", defByService(serviceName, "booking", "driver"))
	dbPass := getEnv("DB_PASSWORD", dbUser)
	dbName := getEnv("DB_NAME", dbUser)

	kBrokers := getEnv("KAFKA_BROKERS", "redpanda:9092")
	tCreated := getEnv("TOPIC_BOOKING_CREATED", "booking.created")
	tAccepted := getEnv("TOPIC_BOOKING_ACCEPTED", "booking.accepted")
	cgJobs := getEnv("CONSUMER_GROUP_JOBS", "driver_svc.jobs")

	return Config{
		ServiceName:          serviceName,
		HTTPPort:             port,
		GracefulTimeout:      time.Duration(gt) * time.Second,
		LogLevel:             logLevel,
		DBHost:               dbHost,
		DBPort:               dbPort,
		DBUser:               dbUser,
		DBPassword:           dbPass,
		DBName:               dbName,
		KafkaBrokers:         kBrokers,
		TopicBookingCreated:  tCreated,
		TopicBookingAccepted: tAccepted,
		ConsumerGroupJobs:    cgJobs,
	}
}

func (c Config) Addr() string { return ":" + c.HTTPPort }

func defByService(name, bookingDefault, driverDefault string) string {
	switch name {
	case "booking_svc":
		return bookingDefault
	case "driver_svc":
		return driverDefault
	default:
		return bookingDefault
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
