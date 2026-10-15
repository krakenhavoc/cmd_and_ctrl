package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reconnaissance Mission — Enchantment {2}{U}{U} (EDHREC top 1,000, #3654):
//
//	"Whenever a creature you control deals combat damage to a player,
//	 you may draw a card.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Bident of Thassa's draw trigger: one event per creature, so two
// creatures that connect are two triggers and two prompts. Cycling is
// the shared hand ability (cycling.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11bf211c-95a3-4e7f-bd08-13f979d54e73",
		Name:         "Reconnaissance Mission",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Reconnaissance Mission — draw a card", Do(DrawCards{N: 1})), "Reconnaissance Mission — draw a card?"),
		},
	})
}
