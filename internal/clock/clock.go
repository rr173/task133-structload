package clock

import (
	"sync"
	"time"
)

// Clock abstracts wall-clock time so business logic and selfcheck stay deterministic.
type Clock interface {
	Now() time.Time
}

// Real returns the actual wall clock.
type Real struct{}

func (Real) Now() time.Time { return time.Now().UTC() }

// Fake is a deterministic clock for tests and selfcheck. Advance moves it forward.
type Fake struct {
	mu  sync.Mutex
	now time.Time
}

func NewFake(start time.Time) *Fake {
	return &Fake{now: start.UTC()}
}

func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *Fake) Advance(d time.Duration) time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
	return f.now
}

func (f *Fake) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = t.UTC()
}
