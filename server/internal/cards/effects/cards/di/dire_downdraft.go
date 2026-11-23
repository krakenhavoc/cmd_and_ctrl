package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Downdraft — Instant {3}{U} (Bloomburrow, #4467):
//
//	"This spell costs {1} less to cast if it targets an attacking or
//	 tapped creature.
//	 Target creature's owner puts it on their choice of the top or bottom
//	 of their library."
//
// Run Behind with a wider discount: CostsLessIfItTargets over a creature
// that is attacking or tapped, read from the announced target as the
// cost is determined (CR 601.2f). The placement is the OWNER's choice,
// not the caster's (ownerTucksFirstTargetTopOrBottom). A creature that
// is gone or no longer a legal target does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a45065e9-951a-4c6d-80b6-bd86a49f7439",
		Name:         "Dire Downdraft",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets an attacking or tapped creature.",
				effects.And(effects.Creature(), effects.Or(effects.IsAttackingCreature(), effects.TappedPermanent()))),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Dire Downdraft — put it on your choice of the top or bottom of your library")
		},
	})
}
