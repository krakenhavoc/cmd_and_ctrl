package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eternity Snare — Enchantment — Aura {5}{U} (Jumpstart 2022):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Charmed Sleep's untap-step hold, which lasts only while the Aura enchants
// the creature, with a draw on entry instead of a tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "b9bd1833-d4f9-46c0-bb94-331c456193fc",
		Name:                  "Eternity Snare",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm, Draws: 1},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eternity Snare — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
