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
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g) && b25DamageDealtByCounteredCreature(ev, g)
			}, "Bred for the Hunt — draw a card", Do(DrawCards{N: 1})), "Bred for the Hunt: draw a card?"),
		},
	})
}
