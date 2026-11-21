package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marut — Artifact Creature — Construct {8}, 7/7 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Trample
//	 When this creature enters, if mana from a Treasure was spent to cast it,
//	 create a Treasure token for each mana from a Treasure spent to cast it."
//
// The intervening "if" (CR 603.4) and the count both read the mana the Marut
// was cast with, recorded as it was spent (Card.ManaSpentToCast): each mana
// that came from a Treasure is one new Treasure, so three Treasures
// sacrificed for mana make three. A Marut put onto the battlefield without
// being cast has no mana spent and makes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4ba37a9e-6b80-4099-bd79-736123d0554b",
		Name:            "Marut",
		ReadsManaSpent:  game.ReadsPoolManaOnly,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.ManaSpentToCast().FromTreasure()
			}), "Marut — create a Treasure token for each mana from a Treasure spent to cast it", func(g *game.Game, item *game.StackItem) error {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				n := src.ManaSpentToCast().CountFrom(game.ManaSourceTreasure)
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
