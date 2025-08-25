package handlerhttp

import (
	"net/http"

	"driver_svc/internal/service"

	"github.com/go-chi/chi/v5"
)

type JobsHandler struct {
	svc service.JobsService
}

func NewJobsHandler(svc service.JobsService) *JobsHandler {
	return &JobsHandler{svc: svc}
}

func (h *JobsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/drivers", h.listDrivers)
	r.Get("/jobs", h.listJobs)
	r.Post("/jobs/{booking_id}/accept", h.acceptJob)
}

func (h *JobsHandler) listDrivers(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListDrivers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list drivers")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *JobsHandler) listJobs(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListOpenJobs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *JobsHandler) acceptJob(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "booking_id")
	var req AcceptJobRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.svc.AcceptJob(r.Context(), bookingID, req.DriverID)
	if err == service.ErrDriverNotFound {
		writeError(w, http.StatusNotFound, "driver not found or unavailable")
		return
	}
	if err == service.ErrJobAlreadyTaken {
		writeError(w, http.StatusConflict, "job already taken")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to accept job")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}
