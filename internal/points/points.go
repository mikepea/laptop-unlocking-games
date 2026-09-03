// Package points is the ledger points are written to. Today the only backend
// is the local profile; the interface exists so TaskBank can slot in later
// without any game or launcher code changing.
package points

import (
	"context"
	"errors"
	"time"

	"github.com/mikepea/aido-linux-unlocker/internal/profile"
)

// ErrNotConfigured is returned by a backend that has not been given the
// settings it needs. Callers should treat it as "skip this backend", not as a
// failure worth showing a child.
var ErrNotConfigured = errors.New("points backend not configured")

// Award is one credit of points.
type Award struct {
	GameID string
	Round  string
	Points int
	Reason string
	At     time.Time
}

// Backend records awards and reports a balance.
type Backend interface {
	Name() string
	Award(ctx context.Context, a Award) error
	Balance(ctx context.Context) (int, error)
}

// Local writes points straight into the in-memory profile. The launcher is
// responsible for persisting the profile afterwards.
type Local struct {
	Profile *profile.Profile
}

// NewLocal returns a Local backend over p.
func NewLocal(p *profile.Profile) *Local { return &Local{Profile: p} }

func (l *Local) Name() string { return "local" }

func (l *Local) Award(_ context.Context, a Award) error {
	if l.Profile == nil {
		return ErrNotConfigured
	}
	l.Profile.AwardPoints(a.Points)
	return nil
}

func (l *Local) Balance(_ context.Context) (int, error) {
	if l.Profile == nil {
		return 0, ErrNotConfigured
	}
	return l.Profile.Balance(), nil
}

// Multi fans an award out to several backends. The first backend is
// authoritative for Balance; the rest are best-effort mirrors, so a TaskBank
// that is down or unconfigured never blocks play.
type Multi struct {
	Backends []Backend
	// OnError, if set, is called for each mirror that fails. Use it to log
	// rather than to interrupt.
	OnError func(b Backend, err error)
}

func (m *Multi) Name() string { return "multi" }

func (m *Multi) Award(ctx context.Context, a Award) error {
	if len(m.Backends) == 0 {
		return ErrNotConfigured
	}
	if err := m.Backends[0].Award(ctx, a); err != nil {
		return err
	}
	for _, b := range m.Backends[1:] {
		if err := b.Award(ctx, a); err != nil && m.OnError != nil {
			m.OnError(b, err)
		}
	}
	return nil
}

func (m *Multi) Balance(ctx context.Context) (int, error) {
	if len(m.Backends) == 0 {
		return 0, ErrNotConfigured
	}
	return m.Backends[0].Balance(ctx)
}
