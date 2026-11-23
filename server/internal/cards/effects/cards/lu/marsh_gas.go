package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marsh Gas — Instant {B} (The Dark, #5638):
//
//	"All creatures get -2/-0 until end of turn."
//
// Hideous Laughter's body with no toughness change. The set is the
// creatures on the battlefield as it resolves; one that arrives later
// this turn is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08ae9f91-48c8-4721-a82d-063defc1b727",
		Name:         "Marsh Gas",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Label: "Marsh Gas — -2/-0"}.Apply(ctx)
		},
	})
}
