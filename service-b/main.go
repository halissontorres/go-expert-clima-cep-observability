package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"

	"github.com/halissontorres/go-expert-clima-cep-observability/service-b/internal/model"
	"github.com/halissontorres/go-expert-clima-cep-observability/service-b/internal/service"
)

var cepRegex = regexp.MustCompile(`^\d{8}$`)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tp, err := initTracer(ctx)
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("error shutting down tracer: %v", err)
		}
	}()

	climaService := service.NewClimaService()

	mux := http.NewServeMux()
	mux.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		handleWeather(w, r, climaService)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Service B starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, otelhttp.NewHandler(mux, "service-b")))
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4317"
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("service-b"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp, nil
}

func handleWeather(w http.ResponseWriter, r *http.Request, cs *service.ClimaService) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	var input model.WeatherInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	cep := strings.TrimSpace(input.CEP)
	if !cepRegex.MatchString(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	city, err := cs.GetLocation(r.Context(), cep)
	if err != nil {
		if err.Error() == "can not find zipcode" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "can not find zipcode"})
			return
		}
		log.Printf("Error fetching location for CEP %s: %v", cep, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	result, err := cs.GetWeather(r.Context(), city)
	if err != nil {
		log.Printf("Error fetching weather for city %s: %v", city, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
