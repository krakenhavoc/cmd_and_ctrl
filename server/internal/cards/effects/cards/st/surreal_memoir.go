package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "fa482ded-b24c-4f9e-948e-d0620e907b8b",
		Name:            "Surreal Memoir",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.GraveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsInstant() })
			pick := ctx.Game.ChooseAtRandomForEffect(effects.RandomDraw(ctx), ids, 1)
			if len(pick) == 0 {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
		},
	})
}
