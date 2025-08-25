package handlerhttp

import (
	"fmt"
)

type AcceptJobRequest struct {
	DriverID string `json:"driver_id"`
}

func (r AcceptJobRequest) Validate() error {
	if r.DriverID == "" {
		return fmt.Errorf("driver_id is required")
	}
	return nil
}
