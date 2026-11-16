package Utils

import (
	"testing"
	"time"
)

func TestDaysInYear(t *testing.T) {
	if d := (TimeExt{T: time.Date(1894, 11, 1, 0, 0, 0, 0, time.UTC)}).DaysInYear(); d != 365 {
		t.Errorf("1894 from T: got %d, want 365", d)
	}
	if d := (TimeExt{Y: 2020}).DaysInYear(); d != 366 {
		t.Errorf("2020 from Y: got %d, want 366", d)
	}
}
