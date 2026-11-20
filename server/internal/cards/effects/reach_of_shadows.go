package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reach of Shadows — Instant {4}{B} (Fate Reforged,
// #5552):
//
//	"Destroy target creature that's one or more colors."
//
// Doom Blade's destroy with the clause "one or more colors": a colorless
// creature (an artifact creature, a face-down 2/2) is not a legal target,
// read as it is targeted and again as the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e01382e0-a685-41c4-9ab6-727ffdbf2e0a",
		Name:         "Reach of Shadows",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature that's one or more colors", Not(Colorless())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
