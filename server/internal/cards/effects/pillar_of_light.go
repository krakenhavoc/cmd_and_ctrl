package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pillar of Light — Instant {2}{W} (Magic 2015, #5556):
//
//	"Exile target creature with toughness 4 or greater."
//
// Murdock's Crusade's first bullet as a whole spell: the toughness is the
// current, layered toughness, judged at announce and again at resolution,
// so a creature shrunk below 4 in response is no longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "643c1ee3-a653-4a6c-b705-a61f3b67e38c",
		Name:         "Pillar of Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
