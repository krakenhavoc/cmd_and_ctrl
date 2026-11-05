package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Command Decision — Sorcery {1}{W} (Star Trek, #4973):
//
//	"Choose one —
//	 • Destroy target creature with power 4 or greater.
//	 • Draw a card."
//
// The power bound is judged at announce and again at resolution (CR
// 608.2b), so a creature shrunk below 4 in response is no longer a
// legal target and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c4afb0d-6d6e-4d21-9fe0-5621cd2fd27c",
		Name:         "Command Decision",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4)),
				DestroyTheModesTarget),
			ModeWithPurpose(ModeDoing("Draw a card.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{N: 1}.Apply(ctx)
				}), game.Purpose{Draws: 1}),
		),
	})
}
