package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surreal Memoir — Sorcery {3}{R}:
//
//	"Return an instant card at random from your graveyard to your hand.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The pick comes from the game's keyed random stream (#744), so it
// replays the same after an undo or a restore. It doesn't target: an
// empty graveyard returns nothing. Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa482ded-b24c-4f9e-948e-d0620e907b8b",
		Name:            "Surreal Memoir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ids := graveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsInstant() })
			pick := ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), ids, 1)
			if len(pick) == 0 {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
		},
	})
}
