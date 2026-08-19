package valueobject

import (
	"testing"
	"time"
)

func TestPeriodOverlap(t *testing.T) {
	a, _ := NewPeriod(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	b, _ := NewPeriod(time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	if !a.Overlaps(b) {
		t.Fatal("expected overlap")
	}
}
func TestInvalidPeriod(t *testing.T) {
	_, e := NewPeriod(time.Now(), time.Now())
	if e == nil {
		t.Fatal("expected error")
	}
}
