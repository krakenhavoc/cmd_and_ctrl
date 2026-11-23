package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High Tide — Instant {U}:
//
//	"Until end of turn, whenever a player taps an Island for mana, that
//	 player adds an additional {U}."
//
// Bubbling Muck's sibling: a triggered mana ability (CR 605.1b) set up for
// the turn (TurnManaTrigger, #2169), so it resolves at once, never uses
// the stack, and is not tied to the instant. It applies to every player's
// Islands, the caster's included, and ends at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc671205-f2fa-454f-9957-921a6069ad53",
		Name:         "High Tide",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TurnManaTrigger{
				Label:   "High Tide — add an additional {U}",
				Subtype: "Island",
				Adds:    "{U}",
			}.Apply(ctx)
		},
	})
}
