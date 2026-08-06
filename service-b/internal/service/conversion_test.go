package service

import (
	"math"
	"testing"
)

func TestConvertTemperatures(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		wantC   float64
		wantF   float64
		wantK   float64
	}{
		{"freezing point", 0, 0, 32, 273},
		{"boiling point", 100, 100, 212, 373},
		{"decimal value", 28.5, 28.5, 83.3, 301.5},
		{"negative value", -10, -10, 14, 263},
		{"body temperature", 36.6, 36.6, 97.88, 309.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTemperatures(tt.celsius, "TestCity")

			if math.Abs(result.TempC-tt.wantC) > 0.1 {
				t.Errorf("TempC = %v, want %v", result.TempC, tt.wantC)
			}
			if math.Abs(result.TempF-tt.wantF) > 0.1 {
				t.Errorf("TempF = %v, want %v", result.TempF, tt.wantF)
			}
			if math.Abs(result.TempK-tt.wantK) > 0.1 {
				t.Errorf("TempK = %v, want %v", result.TempK, tt.wantK)
			}
			if result.City != "TestCity" {
				t.Errorf("City = %v, want TestCity", result.City)
			}
		})
	}
}
