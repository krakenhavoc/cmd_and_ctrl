package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luminous Rebuke — Instant {4}{W} (Foundations, #4465):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Destroy target creature."
//
// Seized from Slumber's text: CostsLessIfItTargets with the
// tapped-creature predicate reads the announced target (CR 601.2f, after
// 601.2c), and destroyTheTargetPermanent destroys whatever the target is
// on resolution, or nothing if it is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d96141af-313f-4ea9-a249-d9eb22cd2431",
		Name:         "Luminous Rebuke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", effects.HobB3TappedCreature()),
		},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
