package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skanos Dragonheart — Legendary Creature — Dragon Ranger {4}{G}, 4/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Whenever Skanos Dragonheart attacks, it gets +X/+X until end of turn,
//	 where X is the greatest power among other Dragons you control and Dragon
//	 cards in your graveyard.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// X is read as the trigger resolves (CR 608.2h): the current power of each other
// Dragon on the battlefield and the printed power of each Dragon card in the
// graveyard. The pump is pinned to this object, so a Skanos that has left and
// come back is not pumped (CR 400.7). Choose a Background is a deck-building
// rule the deck validator applies (deck/partner.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "250052a7-1ba5-45d2-aa32-b787fd637fa5",
		Name:         "Skanos Dragonheart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Skanos Dragonheart — it gets +X/+X until end of turn, where X is the greatest power among other Dragons you control and Dragon cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.GapB217GreatestDragonPower(g, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Dragon"))
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: x, Toughness: x, Label: "Skanos Dragonheart — +X/+X until end of turn"}.Apply(ctx)
				}),
		},
	})
}
