package handler

import (
	"encoding/json"
	"net/http"

	"pos-multi-branch/backend/internal/electric"
)

// ElectricHandler serves ElectricSQL shape management endpoints.
type ElectricHandler struct{}

// NewElectricHandler creates a new ElectricHandler.
func NewElectricHandler() *ElectricHandler {
	return &ElectricHandler{}
}

// Shapes returns the status of all registered ElectricSQL shapes.
// GET /api/v1/electric/shapes
func (h *ElectricHandler) Shapes(w http.ResponseWriter, r *http.Request) {
	statuses := electric.GetShapeStatuses()
	if statuses == nil {
		statuses = []electric.ShapeStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"shapes": statuses,
	})
}
