package main

import "testing"

func TestCompareVersionsValidInput(t *testing.T) {
	tests := []struct {
		a, b  string
		order int
	}{
		{"1.0", "1.0", 0},
		{"0.11", "11.0", -1},
		{"1.1", "1.0", 1},
	}
	for _, tt := range tests {
		o, err := compareVersions(tt.a, tt.b)
		if err != nil {
			t.Errorf("got unexpected error: %v", err)
		}
		if o != tt.order {
			t.Errorf("wrong order: got %d, expected %d", o, tt.order)
		}
	}
}

func TestCompareVersionsInvalidInput(t *testing.T) {
	_, err := compareVersions("--help", "hai")
	if err == nil {
		t.Errorf("expected an error")
	}
}
