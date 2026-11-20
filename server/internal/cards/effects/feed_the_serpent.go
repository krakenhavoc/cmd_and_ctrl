package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feed the Serpent — Instant {2}{B}{B} (Kaldheim, #5505):
//
//	"Exile target creature or planeswalker."
//
// Wander Off's exile (exileTheFirstTargetIfLegal) on Never's "creature
// or planeswalker" target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad404b46-dee7-41bb-969d-35b4cbaadece",
		Name:         "Feed the Serpent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
