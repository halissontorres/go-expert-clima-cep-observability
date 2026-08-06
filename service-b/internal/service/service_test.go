package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/halissontorres/go-expert-clima-cep-observability/service-b/internal/model"
)

func TestGetLocation(t *testing.T) {
	t.Run("successful location lookup", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(model.ViaCEPResponse{
				Localidade: "São Paulo",
			})
		}))
		defer server.Close()

		s := NewClimaServiceWithURLs(server.URL, "")
		city, err := s.GetLocation(context.Background(), "01001000")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if city != "São Paulo" {
			t.Errorf("got city %q, want %q", city, "São Paulo")
		}
	})

	t.Run("CEP not found returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(model.ViaCEPResponse{
				Erro: "true",
			})
		}))
		defer server.Close()

		s := NewClimaServiceWithURLs(server.URL, "")
		_, err := s.GetLocation(context.Background(), "00000000")
		if err == nil {
			t.Error("expected error for not found CEP")
		}
		if err.Error() != "can not find zipcode" {
			t.Errorf("got error %q, want %q", err.Error(), "can not find zipcode")
		}
	})

	t.Run("empty localidade returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(model.ViaCEPResponse{})
		}))
		defer server.Close()

		s := NewClimaServiceWithURLs(server.URL, "")
		_, err := s.GetLocation(context.Background(), "00000000")
		if err == nil {
			t.Error("expected error for empty localidade")
		}
	})
}

func TestGetWeather(t *testing.T) {
	t.Run("successful weather lookup", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(model.WeatherAPIResponse{
				Current: struct {
					TempC float64 `json:"temp_c"`
				}{TempC: 28.5},
			})
		}))
		defer server.Close()

		os.Setenv("WEATHER_API_KEY", "test-key")
		defer os.Unsetenv("WEATHER_API_KEY")

		s := NewClimaServiceWithURLs("", server.URL)
		result, err := s.GetWeather(context.Background(), "São Paulo")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.TempC != 28.5 {
			t.Errorf("TempC = %v, want 28.5", result.TempC)
		}
		if result.TempF != 83.3 {
			t.Errorf("TempF = %v, want 83.3", result.TempF)
		}
		if result.TempK != 301.5 {
			t.Errorf("TempK = %v, want 301.5", result.TempK)
		}
	})

	t.Run("missing API key returns error", func(t *testing.T) {
		os.Unsetenv("WEATHER_API_KEY")

		s := NewClimaService()
		_, err := s.GetWeather(context.Background(), "São Paulo")
		if err == nil {
			t.Error("expected error for missing API key")
		}
	})

	t.Run("API error status returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		os.Setenv("WEATHER_API_KEY", "test-key")
		defer os.Unsetenv("WEATHER_API_KEY")

		s := NewClimaServiceWithURLs("", server.URL)
		_, err := s.GetWeather(context.Background(), "InvalidCity")
		if err == nil {
			t.Error("expected error for API failure")
		}
	})
}
