package points

import (
	"context"
)

// TaskBank is the placeholder for the chore-and-reward ledger that will
// eventually be the real source of truth for points.
//
// It is deliberately inert: with no BaseURL it reports ErrNotConfigured and the
// launcher falls back to the local ledger. When TaskBank exists, fill in the
// two methods and add it to a points.Multi behind Local — the rest of the app
// does not need to know.
type TaskBank struct {
	// BaseURL of the TaskBank API, e.g. https://taskbank.lan/api.
	BaseURL string
	// Token authenticates this laptop to TaskBank.
	Token string
	// PlayerID is the TaskBank account points are credited to.
	PlayerID string
}

func (t *TaskBank) Name() string { return "taskbank" }

func (t *TaskBank) configured() bool {
	return t.BaseURL != "" && t.Token != "" && t.PlayerID != ""
}

func (t *TaskBank) Award(_ context.Context, _ Award) error {
	if !t.configured() {
		return ErrNotConfigured
	}
	// TODO(taskbank): POST {BaseURL}/players/{PlayerID}/awards with an
	// idempotency key so a retry after a flaky link cannot double-credit.
	return ErrNotConfigured
}

func (t *TaskBank) Balance(_ context.Context) (int, error) {
	if !t.configured() {
		return 0, ErrNotConfigured
	}
	// TODO(taskbank): GET {BaseURL}/players/{PlayerID}/balance.
	return 0, ErrNotConfigured
}
