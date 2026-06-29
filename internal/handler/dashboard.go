package handler

import (
	"net/http"
	"time"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
)

// DashboardHandler handles dashboard-related endpoints.
type DashboardHandler struct{}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// DashboardStats returns summary statistics for the dashboard.
// GET /api/v1/dashboard/stats?branch_id=
func (h *DashboardHandler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	branchIDStr := r.URL.Query().Get("branch_id")
	if branchIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch_id"})
		return
	}

	stats, err := repository.GetDashboardStats(branchID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// SalesChart returns daily sales data for the dashboard chart.
// GET /api/v1/dashboard/sales-chart?start=&end=&branch_id=
func (h *DashboardHandler) SalesChart(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	branchIDStr := q.Get("branch_id")
	if branchIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
		return
	}

	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid branch_id"})
		return
	}

	start, end, err := parseTimeRange(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Ensure end is at end of day
	end = end.Truncate(24 * time.Hour).Add(24 * time.Hour)

	data, err := repository.GetSalesChartData(branchID, start, end)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if data.Rows == nil {
		data.Rows = []model.SalesChartRow{}
	}

	writeJSON(w, http.StatusOK, data)
}
