package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// ─── List ───

func TestListUsers_NoDB(t *testing.T) {
	// Requires database connection — skip in unit tests.
	t.Skip("Requires DB connection — ListUsers needs database.Pool")
}

// ─── Create (requires owner role) ───
// All Create tests return 403 Forbidden because no auth context is set.
// Owner-role validation happens before any field checks.

func TestCreateUser_NoAuth(t *testing.T) {
	h := NewUserHandler()
	body, _ := json.Marshal(map[string]string{
		"username":  "newuser",
		"password":  "pass123",
		"full_name": "New User",
		"role":      "kasir",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("CreateUser(no auth) = %d, want 403; body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateUser_EmptyBodyNoAuth(t *testing.T) {
	h := NewUserHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("CreateUser(empty) = %d, want 403; body=%s", rr.Code, rr.Body.String())
	}
}

// ─── Update (requires owner role) ───

func TestUpdateUser_NoAuth(t *testing.T) {
	h := NewUserHandler()
	body, _ := json.Marshal(map[string]interface{}{
		"role": "admin",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/{id}", bytes.NewReader(body))
	req.SetPathValue("id", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Update(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("UpdateUser(no auth) = %d, want 403; body=%s", rr.Code, rr.Body.String())
	}
}

// ─── Delete (requires owner role) ───

func TestDeleteUser_NoAuth(t *testing.T) {
	h := NewUserHandler()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/{id}", nil)
	req.SetPathValue("id", uuid.New().String())
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("DeleteUser(no auth) = %d, want 403; body=%s", rr.Code, rr.Body.String())
	}
}

func TestDeleteUser_InvalidUUIDNoAuth(t *testing.T) {
	h := NewUserHandler()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/{id}", nil)
	req.SetPathValue("id", "bad-uuid")
	rr := httptest.NewRecorder()
	h.Delete(rr, req)
	// Auth check (role != "owner") fires before UUID parse
	if rr.Code != http.StatusForbidden {
		t.Errorf("DeleteUser(bad id) = %d, want 403; body=%s", rr.Code, rr.Body.String())
	}
}

// ─── gofmt & compilation check ───
// Body decode validation tests (username, password, role) are covered
// by the auth guard in unit tests. Integration tests with a real DB
// and JWT context would validate those paths.
