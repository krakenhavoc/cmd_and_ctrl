package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Live Fast — Sorcery {2}{B}:
//
//	"You draw two cards, lose 2 life, and get {E}{E} (two energy
//	 counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04d02394-e17f-43cd-932f-f58e8cc56629",
		Name:         "Live Fast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Energy: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (DrawCards{Player: controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -2); err != nil {
				return err
			}
			return GetEnergy{N: 2}.Apply(ctx)
		},
	})
}
