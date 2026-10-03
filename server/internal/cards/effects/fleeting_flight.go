package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fleeting Flight — Instant {W}:
//
//	"Put a +1/+1 counter on target creature. It gains flying until end of
//	 turn. Prevent all combat damage that would be dealt to it this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the counter and the keyword, then
// a combat-only not-one-use shield pinned to the creature, so its
// non-combat damage is still dealt.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6336401f-4e2d-4ebe-8c5b-24aa6f516abf",
		Name:         "Fleeting Flight",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return thenShieldTheTarget(ctx, true, func(id uuid.UUID) error {
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Fleeting Flight — flying"}.Apply(ctx)
			})
		},
	})
}
