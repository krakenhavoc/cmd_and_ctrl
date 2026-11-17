package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Refresh — Instant {2}{G} (Odyssey, #5613):
//
//	"Regenerate target creature.
//	 Draw a card."
//
// Yavimaya Hollow's regeneration shield (CR 701.19a) from a spell, then
// a card for its caster, in the printed order. Its one target gone in
// response, the spell does nothing at all, the draw included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a4c6643-c48b-4e83-b0c4-96873c13408c",
		Name:         "Refresh",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive, Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := regenerateTheTargetPermanent(ctx.Game, item); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
