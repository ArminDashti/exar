package jalali_test

import (
	"testing"

	"github.com/armin/expenses/backend/internal/jalali"
)

func TestToGregorian(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"1405-06-10", "2026-09-01", false},
		{"1404-01-01", "2025-03-21", false},
		{"invalid", "", true},
		{"1405-13-01", "", true},
	}

	for _, tt := range tests {
		got, err := jalali.ToGregorian(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ToGregorian(%q) expected error, got %q", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ToGregorian(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("ToGregorian(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
