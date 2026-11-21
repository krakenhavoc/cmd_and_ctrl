package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beacon of Destruction — Instant {3}{R}{R} (Fifth Dawn, #5604):
//
//	"Beacon of Destruction deals 5 damage to any target. Shuffle Beacon
//	 of Destruction into its owner's library."
//
// Lightning Bolt's body for five, then the Beacons' tail: the spell goes
// into its owner's library and that library is shuffled. With its one
// target gone the spell does nothing (CR 608.2b) and goes to the
// graveyard, as its ruling says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15feb59b-bbbe-462a-8d38-d20735bbc2f3",
		Name:         "Beacon of Destruction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			// The shuffle waits for a prompt the damage may ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error { return effects.DamageToFirstTargetStep(5)(ctx.Item, ctx) }),
			effects.Instruction(effects.ShuffleThisSpellIntoItsOwnersLibrary),
		),
	})
}
