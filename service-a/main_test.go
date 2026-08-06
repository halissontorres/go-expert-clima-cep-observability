package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCEPValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid CEP with 8 digits",
			body:       `{"cep": "01001000"}`,
			wantStatus: http.StatusInternalServerError, // service B is not available
		},
		{
			name:       "invalid CEP - too short",
			body:       `{"cep": "12345"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantError:  "invalid zipcode",
		},
		{
			name:       "invalid CEP - letters",
			body:       `{"cep": "01001a00"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantError:  "invalid zipcode",
		},
		{
			name:       "invalid CEP - empty",
			body:       `{"cep": ""}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantError:  "invalid zipcode",
		},
		{
			name:       "invalid CEP - special chars",
			body:       `{"cep": "01001-000"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantError:  "invalid zipcode",
		},
		{
			name:       "invalid JSON body",
			body:       `not json`,
			wantStatus: http.StatusUnprocessableEntity,
			wantError:  "invalid zipcode",
		},
		{
			name:       "GET method not allowed",
			body:       ``,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.name == "GET method not allowed" {
				req = httptest.NewRequest(http.MethodGet, "/", nil)
			} else {
				req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.body)))
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			handleCEP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantError != "" {
				var body map[string]string
				json.NewDecoder(w.Body).Decode(&body)
				if body["error"] != tt.wantError {
					t.Errorf("got error %q, want %q", body["error"], tt.wantError)
				}
			}
		})
	}
}

func TestHandleCEPContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"cep": "01001000"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleCEP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
}
