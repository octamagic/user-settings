//go:generate mockgen -source=clock.go -destination=mock_clock.go -package=domain
package domain

import "time"

type Clock interface {
	Now() time.Time
}
type DefaultClock struct{}

func (c *DefaultClock) Now() time.Time {
	return time.Now()
}
