package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero's Demise — Instant {1}{B} (Betrayers of Kamigawa, #5602):
//
//	"Destroy target legendary creature."
//
// Doom Blade's shape with the legendary supertype as the target's
// condition, judged at announce and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69ec0b59-a0aa-4878-a3b3-130cea18fee0",
		Name:         "Hero's Demise",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target legendary creature", Legendary()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
