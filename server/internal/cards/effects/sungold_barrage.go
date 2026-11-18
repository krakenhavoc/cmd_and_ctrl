package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sungold Barrage — Instant {2}{W} (Innistrad: Midnight Hunt, #5497):
//
//	"Destroy target creature with toughness 4 or greater."
//
// Murdock's Crusade's toughness predicate over Doom Blade's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a33d0ea-9b19-4fde-8e7a-2a9d17f716ea",
		Name:         "Sungold Barrage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    s82MidDestroyTheTarget,
	})
}
