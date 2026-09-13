package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night's Whisper — Sorcery {1}{B}:
//
//	"You draw two cards and you lose 2 life."
//
// Sign in Blood without the target — the caster always pays and
// always draws, which is why it costs one less. Same DrawCards +
// ChangePlayerLife composition, aimed at the controller.
func init() {
	Register(Spec{
		OracleID:     "7ffae8f8-3006-4969-a339-6d30678f87ea",
		Name:         "Night's Whisper",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (DrawCards{Player: controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -2)
		},
	})
}
