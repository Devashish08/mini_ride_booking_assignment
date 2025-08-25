package events

type BookingAccepted struct {
	BookingID  string `json:"booking_id"`
	DriverID   string `json:"driver_id"`
	RideStatus string `json:"ride_status"` // "Accepted"
}
