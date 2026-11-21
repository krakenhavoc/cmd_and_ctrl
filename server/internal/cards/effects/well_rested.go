package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Well Rested — Enchantment — Aura {1}{G} (Fallout):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever this creature becomes untapped, put two
//	 +1/+1 counters on it, then you gain 2 life and draw a card. This
//	 ability triggers only once each turn.""
//
// Fishing Pole's untap watch on a trigger the Aura grants the creature:
// "you" is the creature's controller and "it" the creature. The once-a-turn
// limit is the engine's record (TriggersOnlyOnceEachTurn), kept per object.
//
// No simplification.
const wellRestedGrant = "well-rested/untap-counters"

func init() {
	Register(Spec{
		OracleID:     "f8cebb33-9235-4f63-8171-15af8b8e1436",
		Name:         "Well Rested",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Draws: 1},
		Grants: []AbilityGrant{{
			Key: wellRestedGrant,
			Triggered: []game.TriggeredAbility{
				TriggersOnlyOnceEachTurn(On(game.EventUntapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventUntapCard && ev.CardID == source.InstanceID
				}, "Well Rested — put two +1/+1 counters on this creature, then you gain 2 life and draw a card",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}, GainLife{Amount: 2}, DrawCards{N: 1}))),
			},
			Text: "Whenever this creature becomes untapped, put two +1/+1 counters on it, then you gain 2 life and draw a card. This ability triggers only once each turn.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(wellRestedGrant)},
	})
}
