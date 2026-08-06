package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/halissontorres/go-expert-clima-cep-observability/service-b/internal/model"
)

var cepRegex = regexp.MustCompile(`^\d{8}$`)

func ValidateCEP(cep string) bool {
	return cepRegex.MatchString(strings.TrimSpace(cep))
}

type ClimaService struct {
	httpClient *http.Client
	viaCEPURL  string
	weatherURL string
}

func NewClimaService() *ClimaService {
	return &ClimaService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		viaCEPURL:  "https://viacep.com.br",
		weatherURL: "https://api.weatherapi.com",
	}
}

func NewClimaServiceWithURLs(viaCEPURL, weatherURL string) *ClimaService {
	return &ClimaService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		viaCEPURL:  viaCEPURL,
		weatherURL: weatherURL,
	}
}

func (s *ClimaService) GetLocation(ctx context.Context, cep string) (string, error) {
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "viacep-get-location")
	defer span.End()

	span.SetAttributes(attribute.String("cep", cep))

	u := fmt.Sprintf("%s/ws/%s/json/", s.viaCEPURL, url.PathEscape(cep))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "viacep request failed")
		return "", err
	}
	defer resp.Body.Close()

	var viaCEP model.ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCEP); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode response")
		return "", err
	}

	if viaCEP.Erro == "true" || viaCEP.Localidade == "" {
		span.SetStatus(codes.Error, "can not find zipcode")
		return "", fmt.Errorf("can not find zipcode")
	}

	span.SetAttributes(attribute.String("city", viaCEP.Localidade))
	return viaCEP.Localidade, nil
}

func (s *ClimaService) GetWeather(ctx context.Context, city string) (*model.WeatherOutput, error) {
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "weatherapi-get-temperature")
	defer span.End()

	span.SetAttributes(attribute.String("city", city))

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		err := fmt.Errorf("WEATHER_API_KEY environment variable is not set")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing API key")
		return nil, err
	}

	u := fmt.Sprintf(
		"%s/v1/current.json?key=%s&q=%s&aqi=no",
		s.weatherURL,
		url.QueryEscape(apiKey),
		url.QueryEscape(city),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "weather API request failed")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("weather API returned status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "weather API error")
		return nil, err
	}

	var weather model.WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to decode response")
		return nil, err
	}

	result := convertTemperatures(weather.Current.TempC, city)
	span.SetAttributes(
		attribute.Float64("temp_c", result.TempC),
		attribute.Float64("temp_f", result.TempF),
		attribute.Float64("temp_k", result.TempK),
	)
	return result, nil
}

func convertTemperatures(celsius float64, city string) *model.WeatherOutput {
	return &model.WeatherOutput{
		City:  city,
		TempC: math.Round(celsius*100) / 100,
		TempF: math.Round((celsius*1.8+32)*100) / 100,
		TempK: math.Round((celsius+273)*100) / 100,
	}
}
