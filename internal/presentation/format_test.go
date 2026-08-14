package presentation

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{duration: 750 * time.Millisecond, want: "750ms"},
		{duration: 12449 * time.Millisecond, want: "12.4s"},
	}
	for _, test := range tests {
		if got := FormatDuration(test.duration); got != test.want {
			t.Errorf("FormatDuration(%s) = %q, want %q", test.duration, got, test.want)
		}
	}
}
