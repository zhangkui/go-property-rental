package valueobject

import "fmt"

type Money int64

func (m Money) String() string    { return fmt.Sprintf("%d", m) }
func (m Money) Add(v Money) Money { return m + v }
func (m Money) Sub(v Money) Money { return m - v }
func (m Money) NonNegative() bool { return m >= 0 }
