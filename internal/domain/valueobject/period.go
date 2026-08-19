package valueobject

import (
	"errors"
	"time"
)

type Period struct{ Start, End time.Time }

func NewPeriod(s, e time.Time) (Period, error) {
	if !s.Before(e) {
		return Period{}, errors.New("period end must be after start")
	}
	return Period{s, e}, nil
}
func (p Period) Overlaps(o Period) bool { return p.Start.Before(o.End) && o.Start.Before(p.End) }
