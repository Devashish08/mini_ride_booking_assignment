package handlerhttp

import (
	"net/http"

	"booking_svc/internal/service"

	"github.com/go-chi/chi/v5"
)

type BookingHandler struct {
	svc service.BookingService
}

func NewBookingHandler(svc service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// RegisterRoutes attaches endpoints to the provided router.
// We’ll call this from the HTTP server wiring in the next step.
func (h *BookingHandler) RegisterRoutes(r chi.Router) {
	r.Post("/bookings", h.createBooking)
	r.Get("/bookings", h.listBookings)
}

func (h *BookingHandler) createBooking(w http.ResponseWriter, r *http.Request) {
	var req CreateBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.svc.CreateBooking(r.Context(), service.CreateBookingInput{
		PickupLoc: req.PickupLoc,
		Dropoff:   req.Dropoff,
		Price:     req.Price,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create booking")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *BookingHandler) listBookings(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListBookings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list bookings")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
