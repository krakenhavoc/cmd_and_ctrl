package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treasonous Ogre — Creature — Ogre Shaman {3}{R}, 2/3 (Secret Lair Drop):
//
//	"Dethrone (Whenever this creature attacks the player with the most life
//	 or tied for most life, put a +1/+1 counter on it.)
//	 Pay 3 life: Add {R}."
//
// Dethrone is the shared GapB201Dethrone trigger (CR 702.105a): the defender
// is judged as the attack is declared, so an attack at a planeswalker never
// triggers it. "Pay 3 life: Add {R}" is a mana ability whose cost is only
// life (CR 605.1a, 118.4): it can be activated with nothing to tap, and
// only while the player has 3 or more life to pay (CR 119.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "826d4279-1576-4898-a1c2-26fd547fb0d0",
		Name:         "Treasonous Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Treasonous Ogre — dethrone: put a +1/+1 counter on it"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Life: 3},
			Produced: "{R}",
			Label:    "Pay 3 life: Add {R}",
		}},
	})
}
