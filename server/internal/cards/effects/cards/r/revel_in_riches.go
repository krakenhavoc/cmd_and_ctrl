package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revel in Riches — Enchantment {4}{B} (Secret Lair Drop):
//
//	"Whenever a creature an opponent controls dies, create a Treasure token.
//	 At the beginning of your upkeep, if you control ten or more Treasures,
//	 you win the game."
//
// The first is the dies trigger (tokens included, last-known control). The
// second is an intervening "if" (CR 603.4): checked as the upkeep begins and
// again as the trigger resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc3fcd26-c088-4335-9a35-0e11019e3d5b",
		Name:         "Revel in Riches",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OldB11OpponentCreatureDied(ev, source, g)
			}, "Revel in Riches — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			effects.InterveningIfYou("you control ten or more Treasures", func(g *game.Game, you uuid.UUID) bool {
				return effects.CountControlled(g, you, func(c game.Card) bool { return c.HasSubtype("Treasure") }) >= 10
			}, effects.AtYourUpkeep("Revel in Riches — you win the game", func(g *game.Game, item *game.StackItem) error {
				return effects.WinTheGame{Player: item.Controller}.Apply(effects.NewContext(g, item))
			})),
		},
	})
}
