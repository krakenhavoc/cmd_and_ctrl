package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shoulder to Shoulder — Sorcery {2}{W} (Foundations Jumpstart):
//
//	"Support 2. (Put a +1/+1 counter on each of up to two target
//	 creatures.)
//	 Draw a card."
//
// Support as a spell (CR 701.41a: any creatures, "up to"), then the
// caster draws. With no target chosen the spell still resolves and draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ffa7cc8-4774-430d-b898-5bd26366e124",
		Name:         "Shoulder to Shoulder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      SupportTargets(2, false),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := SupportEffect(ctx.Game, item); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
