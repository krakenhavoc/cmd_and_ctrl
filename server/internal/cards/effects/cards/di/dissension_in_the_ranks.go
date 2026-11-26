package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dissension in the Ranks — Instant {3}{R}{R} (Shadows over Innistrad, S82
// OLD):
//
//	"Target blocking creature fights another target blocking creature."
//
// Blood Feud's two clauses with the blocking filter on each; the fight happens
// only while both are still legal targets (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "717266b6-2bbc-4a04-8f30-e3eaf0283acd",
		Name:         "Dissension in the Ranks",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			effects.Distinct(effects.TargetCreature("another target blocking creature", effects.BlockingCreature())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SlotZeroFightsSlotOne(ctx.Game, item)
		},
	})
}
