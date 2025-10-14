package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI_checkHandler(t *testing.T) {
	t.Run("valid comment", func(t *testing.T) {
		body := []byte(`{"content": "This is a clean comment"}`)
		req := httptest.NewRequest(http.MethodPost, "/check?request_id=abc123", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		api := New()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var resp map[string]string
		err := json.NewDecoder(rr.Body).Decode(&resp)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["status"] != "ok" {
			t.Errorf("expected status ok, got %v", resp["status"])
		}
	})

	t.Run("banned word", func(t *testing.T) {
		body := []byte(`{"content": "qwerty forbidden"}`)
		req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		api := New()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		body := []byte(`{"content": "missing quote}`)
		req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		api := New()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}
