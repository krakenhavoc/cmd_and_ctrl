package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wisdom of Ages — Sorcery {4}{U}{U}{U}:
//
//	"Return all instant and sorcery cards from your graveyard to your
//	 hand. You have no maximum hand size for the rest of the game.
//	 Exile Wisdom of Ages."
//
// Praetor's Counsel's shape with a type filter: the graveyard is read once
// (returnWholeGraveyardToHand), the spell exiles itself while it is still on
// the stack, and the no-maximum clause is the player-level grant
// (Game.SetMaxHandSizeForEffect) that outlives the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9722ec3-aa5e-4c69-8ff0-cef1b1839fa4",
		Name:         "Wisdom of Ages",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := returnWholeGraveyardToHand(ctx, func(c game.Card) bool {
				return c.IsInstant() || c.IsSorcery()
			}); err != nil {
				return err
			}
			if err := ctx.Game.SetMaxHandSizeForEffect(ctx.Controller(), game.NoMaxHandSize); err != nil {
				return err
			}
			return (ExileTarget{Target: ctx.Source()}).Apply(ctx)
		},
	})
}
