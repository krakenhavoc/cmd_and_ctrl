package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urborg Justice — Instant {B}{B} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Target opponent sacrifices a creature of their choice for each creature put into your graveyard
//	 from the battlefield this turn."
//
// The number is read as the spell resolves from the turn's zone changes: creature cards the caster owns that went from the battlefield to a
// graveyard (CR 400.3), counting one that has since left it, and tokens, which are put into the graveyard before they cease to exist
// (CR 111.7). The opponent then sacrifices that many creatures, choosing each (CR 701.21); with fewer creatures they sacrifice what they have.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8e2dea9-0e96-4ce4-a070-9893f44cf0cc",
		Name:         "Urborg Justice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB779UrborgJusticeSacrifices(ctx)
		},
	})
}
