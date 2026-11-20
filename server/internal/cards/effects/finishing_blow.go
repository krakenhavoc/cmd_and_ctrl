package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Finishing Blow — Instant {4}{B} (Core Set 2021, #5511):
//
//	"Destroy target creature or planeswalker."
//
// Hero's Downfall's clause over both types; the destroy honours
// indestructible and skips a target that left in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c8a6c58-f4af-4809-af86-a931313fa71e",
		Name:         "Finishing Blow",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
