package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unsubtle Mockery — Instant {2}{R} (Secrets of Strixhaven, #4456):
//
//	"Unsubtle Mockery deals 4 damage to target creature. Surveil 1. (Look at
//	 the top card of your library. You may put it into your graveyard.)"
//
// Two sentences in order: the damage, then the surveil. The surveil waits on
// a prompt the damage asked (a damage replacement, CR 608.2c), and a spell
// whose only target is gone does nothing at all, the surveil included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d059345-1a2f-4ecb-b419-3873ebadf318",
		Name:         "Unsubtle Mockery",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve: ResolveSteps(
			Instruction(socB38DamageToFirstTarget(4)),
			Surveil{N: 1}),
	})
}
