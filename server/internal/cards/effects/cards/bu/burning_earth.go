package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Earth — Enchantment {3}{R} (Magic 2014, S82 old sets):
//
//	"Whenever a player taps a nonbasic land for mana, this enchantment deals 1
//	 damage to that player."
//
// Manabarbs' trigger narrowed to a land without the basic supertype (CR
// 205.4c, read from the land as it now stands): it watches
// EventManaAbilityActivated, which a manual activation and the auto-tapper both
// emit, and fires for a land that is now tapped, so a mana ability with no {T}
// in its cost is not "tapping for mana". Every player, the controller included;
// the damage is dealt by the enchantment to the player who tapped, read off the
// carried trigger context.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21b8192e-46a2-4a6a-abd5-3a074efa30a9",
		Name:         "Burning Earth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB400NonbasicLandTappedForMana(ev, g)
			},
			Key: "Burning Earth — deal 1 damage to that player",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
