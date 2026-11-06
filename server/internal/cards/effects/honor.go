package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honor — Sorcery {W} (Edge of Eternities, #4461):
//
//	"Put a +1/+1 counter on target creature.
//	 Draw a card."
//
// Hunter's Edge's counter step (counterOnFirstCreatureTarget) followed by
// a draw. With its only target gone the spell doesn't resolve at all, so
// the card isn't drawn either (CR 608.2b), as the printed ruling says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48c0cd06-465d-4901-bf23-948a003afc0b",
		Name:         "Honor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(Instruction(counterOnFirstCreatureTarget), DrawCards{N: 1}),
	})
}
