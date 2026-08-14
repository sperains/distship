package presentation

import (
	"fmt"
	"time"
)

func FormatDuration(duration time.Duration) string {
	if duration < time.Second {
		return fmt.Sprintf("%dms", duration.Milliseconds())
	}
	return duration.Round(100 * time.Millisecond).String()
}
