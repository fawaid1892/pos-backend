package electric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// ShapeConfig defines a table shape for ElectricSQL replication.
type ShapeConfig struct {
	Table      string   `json:"table"`
	PrimaryKey []string `json:"primary_key"`
	Columns    []string `json:"columns"`
}

// ShapeStatus holds the registration status of a shape.
type ShapeStatus struct {
	Table  string `json:"table"`
	Status string `json:"status"` // "registered", "error"
	Error  string `json:"error,omitempty"`
}

var (
	shapeStatuses []ShapeStatus
	statusMu      sync.RWMutex
)

// shapeDefinitions returns the full list of table shapes to register.
func shapeDefinitions() []ShapeConfig {
	return []ShapeConfig{
		{
			Table:      "users",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "username", "full_name", "role", "branch_id", "created_at", "updated_at"},
		},
		{
			Table:      "branches",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "name", "address", "phone", "tax_rate", "is_active", "created_at", "updated_at", "deleted_at"},
		},
		{
			Table:      "categories",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "name", "created_at", "updated_at"},
		},
		{
			Table:      "products",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "category_id", "name", "barcode", "price", "cost_price", "stock", "created_at", "updated_at", "deleted_at"},
		},
		{
			Table:      "transactions",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "branch_id", "user_id", "customer_name", "subtotal", "discount_percent", "discount_amount", "tax_rate", "tax_amount", "total", "cash_amount", "change_amount", "payment_method", "payment_reference", "created_at"},
		},
		{
			Table:      "transaction_items",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "transaction_id", "product_id", "product_name", "quantity", "price", "subtotal"},
		},
		{
			Table:      "branch_products",
			PrimaryKey: []string{"branch_id", "product_id"},
			Columns:    []string{"branch_id", "product_id", "stock_qty", "created_at", "updated_at"},
		},
		{
			Table:      "stock_mutations",
			PrimaryKey: []string{"id"},
			Columns:    []string{"id", "branch_id", "product_id", "type", "qty", "reference_id", "notes", "created_at"},
		},
	}
}

// InitShapes connects to the ElectricSQL sync service, checks health, and
// registers all table shapes so they are replicated via logical replication.
func InitShapes(electricURL string) error {
	statusMu.Lock()
	defer statusMu.Unlock()

	// Reset statuses
	shapeStatuses = nil

	// ─── Health check ───
	if err := healthCheck(electricURL); err != nil {
		return fmt.Errorf("electric health check failed: %w", err)
	}

	// ─── Check existing shapes ───
	if err := listShapes(electricURL); err != nil {
		// Non-fatal: log but continue with registration
		fmt.Printf("[electric] warning: could not list existing shapes: %v\n", err)
	}

	// ─── Register each shape ───
	defs := shapeDefinitions()
	for _, shape := range defs {
		status, err := registerShape(electricURL, shape)
		if err != nil {
			shapeStatuses = append(shapeStatuses, ShapeStatus{
				Table:  shape.Table,
				Status: "error",
				Error:  err.Error(),
			})
			fmt.Printf("[electric] failed to register shape %s: %v\n", shape.Table, err)
			continue
		}
		shapeStatuses = append(shapeStatuses, ShapeStatus{
			Table:  shape.Table,
			Status: status,
		})
		fmt.Printf("[electric] shape %s: %s\n", shape.Table, status)
	}

	return nil
}

// GetShapeStatuses returns a copy of the current shape registration statuses.
func GetShapeStatuses() []ShapeStatus {
	statusMu.RLock()
	defer statusMu.RUnlock()
	result := make([]ShapeStatus, len(shapeStatuses))
	copy(result, shapeStatuses)
	return result
}

// healthCheck hits the ElectricSQL /health endpoint.
func healthCheck(electricURL string) error {
	url := electricURL + "/health"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s returned %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}

// listShapes hits GET /v1/shape to verify the service is ready.
func listShapes(electricURL string) error {
	url := electricURL + "/v1/shape"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s returned %d: %s", url, resp.StatusCode, string(body))
	}
	return nil
}

// registerShape posts a single shape definition to ElectricSQL.
// Returns "registered" on success.
func registerShape(electricURL string, shape ShapeConfig) (string, error) {
	url := electricURL + "/v1/shape/" + shape.Table

	body, err := json.Marshal(shape)
	if err != nil {
		return "", fmt.Errorf("marshal shape %s: %w", shape.Table, err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("POST %s returned %d: %s", url, resp.StatusCode, string(respBody))
	}

	return "registered", nil
}
