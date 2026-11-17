package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "15feb59b-bbbe-462a-8d38-d20735bbc2f3",
		Name:         "Beacon of Destruction",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			// The shuffle waits for a prompt the damage may ask (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error { return damageToFirstTarget(5)(ctx.Item, ctx) }),
			Instruction(shuffleThisSpellIntoItsOwnersLibrary),
		),
	})
}
