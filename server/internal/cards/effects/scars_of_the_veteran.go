package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scars of the Veteran — Instant {4}{W}:
//
//	"You may exile a white card from your hand rather than pay this spell's mana cost.
//	 Prevent the next 7 damage that would be dealt to any target this turn. If it's a creature, put a +0/+1 counter on it for each 1 damage prevented this way at the beginning of the next end step."
//
// ADR 0108 owner decision 2 (#1906): Sacred Boon's shape on any target,
// for 7 (sacred_boon.go). A player is shielded the same way and gets no
// counters. The alternative cost is the pitch (CR 118.9).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5f0ad09f-a9ee-42df-b32d-9d842ce5dd00",
		Name:         "Scars of the Veteran",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		AlternativeCosts: []game.AlternativeCost{
			Pitch(
				"Exile a white card from your hand",
				0,
				CardInYourHand("a white card from your hand", OfColor("W")),
				"a white card from your hand",
			),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return endStepToughnessShield(ctx, 7, "Scars of the Veteran")
		},
	})
}
