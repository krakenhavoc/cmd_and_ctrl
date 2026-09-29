package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corrupted Conviction — Instant {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// Village Rites under a different name and printing — the sacrifice
// is a COST, so the creature dies with Corrupted Conviction still on
// the stack: a Blood Artist trigger drains before the cards are
// drawn, and countering the spell doesn't hand the creature back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "b45e35df-9032-4482-89a6-c7c50c6d0a79",
		Name:           "Corrupted Conviction",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
