package service

import "testing"

func TestValidateCEP(t *testing.T) {
	tests := []struct {
		name string
		cep  string
		want bool
	}{
		{"valid CEP", "01001000", true},
		{"valid CEP with spaces", " 01001000 ", true},
		{"too short", "1234567", false},
		{"too long", "123456789", false},
		{"letters", "01001a00", false},
		{"empty", "", false},
		{"special chars", "01001-000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCEP(tt.cep); got != tt.want {
				t.Errorf("ValidateCEP(%q) = %v, want %v", tt.cep, got, tt.want)
			}
		})
	}
}
