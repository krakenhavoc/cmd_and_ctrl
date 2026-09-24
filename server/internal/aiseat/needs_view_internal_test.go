package aiseat

import (
	"context"
	"testing"
)

// needs_view_internal_test.go pins Runner.needsView (#1261): which
// decision windows may skip building the seat's view. In-package
// because the rule is unexported and small, and a unit test states it
// more exactly than timing a game would.

// blindWrapper wraps a view-blind policy the way a tier or a
// decision-log shim would. It is exactly the case the marker must NOT
// leak through: the wrapper's own Decide may read the view.
type blindWrapper struct{ inner Policy }

func (w blindWrapper) Name() string { return "wrapper" }
func (w blindWrapper) Decide(ctx context.Context, in Input) (Decision, error) {
	return w.inner.Decide(ctx, in)
}
func (w blindWrapper) Unwrap() Policy { return w.inner }

type nopObserver struct{}

func (nopObserver) Observe(DecisionEvent) {}

type concedingRandom struct{ *RandomPolicy }

func (concedingRandom) ShouldConcede(Input) bool { return false }

func TestNeedsView(t *testing.T) {
	random := NewRandomPolicy(nil)
	cases := []struct {
		name   string
		policy Policy
		cfg    Config
		want   bool
	}{
		{"a bare random seat decides without a view", random, Config{}, false},
		{"an observer receives the whole Input", random, Config{Observer: nopObserver{}}, true},
		{"a wrapper may read the view even when what it wraps does not", blindWrapper{random}, Config{}, true},
		{"a conceder reads the view to decide whether to act at all", concedingRandom{random}, Config{}, true},
		{"a policy that makes no claim gets its view", blindWrapper{blindWrapper{random}}.inner, Config{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := &Runner{policy: c.policy, cfg: c.cfg}
			if got := r.needsView(); got != c.want {
				t.Errorf("needsView() = %v, want %v", got, c.want)
			}
		})
	}
}
