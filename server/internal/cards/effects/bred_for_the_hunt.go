package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bred for the Hunt — Enchantment {1}{G}{U} (EDHREC rank 2668):
//
//	"Whenever a creature you control with a +1/+1 counter on it deals
//	 combat damage to a player, you may draw a card."
//
// The Simic counters deck's Coastal Piracy. The trigger is the
// "creature you control deals combat damage to a player" shape
// (combatDamageToPlayerBy — the creature is read live, since combat
// damage lands before the state-based sweep) narrowed to a creature
// carrying at least one +1/+1 counter (b25DamageDealtByCounteredCreature).
// One event per creature, so three countered attackers connecting
// ask three times, as printed; "you may" is the yes/no prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc7cda6f-7e5e-4a56-8d67-ffdea7edf269",
		Name:         "Bred for the Hunt",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g) && b25DamageDealtByCounteredCreature(ev, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Bred for the Hunt: draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bred for the Hunt — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
