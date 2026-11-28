package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Femeref Enchantress — Creature — Human Druid {G}{W}, 1/2 (Visions, S82 OLD
// batch 787, #5631):
//
//	"Whenever an enchantment is put into a graveyard from the battlefield,
//	 draw a card."
//
// Fires for any player's enchantment, token or card, the Enchantress's own
// controller's included, read as the permanent last existed on the battlefield
// (CR 603.10a). One trigger per enchantment, so a wrath of three enchantments
// draws three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b983a46-580f-4a94-8ef3-5253e503148e",
		Name:         "Femeref Enchantress",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftAsType(ev, c, "enchantment")
			},
			Key:     "Femeref Enchantress — draw a card",
			Purpose: game.Purpose{Draws: 1},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
