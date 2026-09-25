package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manabarbs — Enchantment {3}{R} (EDHREC rank 2226):
//
//	"Whenever a player taps a land for mana, this enchantment deals 1
//	 damage to that player."
//
// The symmetrical land tax. The trigger watches
// EventManaAbilityActivated — emitted for a manual activation and
// for the auto-tapper alike, and narrower than "becomes tapped",
// which is the difference between this card and Citadel of Pain —
// and fires when the source is a land that is now tapped
// (b20LandTappedForMana). Every player, the controller included; the
// damage is dealt by the enchantment, so it is noncombat damage from
// a noncreature source, and the victim is the event's Actor, read at
// resolution off the item's carried trigger context
// (item.Trigger.Event.Actor, #1223).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f1afedd-c60f-454f-b84a-c8117aec0128",
		Name:         "Manabarbs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20LandTappedForMana(ev, g)
			},
			Key: "Manabarbs — deal 1 damage to that player",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
