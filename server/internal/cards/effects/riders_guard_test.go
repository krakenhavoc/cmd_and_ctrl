package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// riders_guard_test.go — #4978: Register's guard for the riders (scry,
// learn, discover, amass, the Ring), and WithRiders.
func TestRidersGuardOnSlots(t *testing.T) {
	for _, c := range []struct {
		name string
		slot purposeSlot
		p    game.Purpose
		want string
	}{
		{"scry on the card", purposeOnCard, game.Purpose{Draws: 1, Scry: 2}, ""},
		{"discover on an activated row", purposeOnActivated, game.Purpose{Discover: 4}, ""},
		{"scry on a modal trigger's bullet", purposeOnMode, game.Purpose{Scry: 2}, ""},
		{"the Ring on a spell's mode", purposeOnSpellMode, game.Purpose{RingTempts: 1}, ""},
		{"amass when turned face up", purposeOnFaceUp, game.Purpose{Amass: 1}, ""},
		{"learn on a triggered row", purposeOnTriggered, game.Purpose{Learn: 1}, "on a triggered row"},
		{"a negative scry", purposeOnCard, game.Purpose{Scry: -1}, "negative rider"},
	} {
		t.Run(c.name, func(t *testing.T) {
			msg := refusal(func() { checkPurpose("Test Card", "card", c.slot, c.p) })
			if c.want == "" && msg != "" {
				t.Fatalf("refused: %s", msg)
			}
			if c.want != "" && !strings.Contains(msg, c.want) {
				t.Fatalf("got %q, want a refusal naming %q", msg, c.want)
			}
		})
	}
}

func TestWithRidersAddsOnlyTheRiders(t *testing.T) {
	base := ForTargets(DamageToTarget(0, 4))
	got := WithRiders(base, game.Purpose{Scry: 1, RingTempts: 1, Discover: 3, Draws: 5})
	if got.Scry != 1 || got.RingTempts != 1 || got.Discover != 3 {
		t.Errorf("riders = scry %d, ring %d, discover %d; want 1, 1, 3", got.Scry, got.RingTempts, got.Discover)
	}
	if got.Draws != 0 {
		t.Errorf("WithRiders copied a draw (%d) that is not a rider", got.Draws)
	}
	if got.Targets == nil || len(got.Targets.List()) != 1 {
		t.Errorf("the target entry was lost: %+v", got.Targets)
	}
}
