package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guidance Failure — Instant {3}{U} (Star Trek, #4973):
//
//	"This spell costs {1} less to cast if it targets an attacking
//	 creature.
//	 Target creature or Spacecraft's owner puts it on their choice of the
//	 top or bottom of their library."
//
// Run Behind with a wider target: a creature, or a Spacecraft that is not
// (yet) a creature. The discount asks only about an attacking creature;
// the owner, not the caster, picks the end of the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c18f853-684b-42f6-8bc6-e86596131ab1",
		Name:         "Guidance Failure",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or Spacecraft", effects.Or(effects.Creature(), effects.HasSubtype("Spacecraft"))),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets an attacking creature.",
				effects.And(effects.Creature(), effects.IsAttackingCreature())),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Guidance Failure — put it on your choice of the top or bottom of your library")
		},
	})
}
