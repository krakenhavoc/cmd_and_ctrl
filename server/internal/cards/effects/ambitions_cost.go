package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambition's Cost — Sorcery {3}{B}:
//
//	"You draw three cards and you lose 3 life."
//
// Night's Whisper scaled up. Same DrawCards + ChangePlayerLife
// composition aimed at the controller.
func init() {
	Register(Spec{
		OracleID:     "84de4fec-2f38-4293-93d3-b3882c5aac14",
		Name:         "Ambition's Cost",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (DrawCards{Player: controller, N: 3}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -3)
		},
	})
}
