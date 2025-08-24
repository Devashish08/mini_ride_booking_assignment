package handlerhttp

import (
	"fmt"
	"strings"

	"booking_svc/internal/models"
)

type CreateBookingRequest struct {
	PickupLoc models.Location `json:"pickuploc"`
	Dropoff   models.Location `json:"dropoff"`
	Price     int             `json:"price"`
}

func (r CreateBookingRequest) Validate() error {
	var errs []string

	if !isValidLat(r.PickupLoc.Lat) {
		errs = append(errs, "pickuploc.lat must be between -90 and 90")
	}
	if !isValidLng(r.PickupLoc.Lng) {
		errs = append(errs, "pickuploc.lng must be between -180 and 180")
	}
	if !isValidLat(r.Dropoff.Lat) {
		errs = append(errs, "dropoff.lat must be between -90 and 90")
	}
	if !isValidLng(r.Dropoff.Lng) {
		errs = append(errs, "dropoff.lng must be between -180 and 180")
	}
	if r.Price <= 0 {
		errs = append(errs, "price must be > 0")
	}
	if r.PickupLoc == r.Dropoff {
		errs = append(errs, "pickuploc and dropoff cannot be the same")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

func isValidLat(v float64) bool { return v >= -90 && v <= 90 }
func isValidLng(v float64) bool { return v >= -180 && v <= 180 }
