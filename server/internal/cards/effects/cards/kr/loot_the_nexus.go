package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loot, the Nexus — Legendary Creature — Beast Noble {2}{G}, 2/1:
//
//	"{T}: Choose a color. Add one mana of that color for each different
//	 power among creatures you control."
//
// A mana ability whose amount is computed at activation (Avatar Kyoshi's
// ProducedFunc / ProducedOneColor shape): one colour pick, then N mana of
// it, N being the count of distinct power values (negative values count)
// among the controller's creatures as the ability is activated. Loot is a
// creature with a tap cost, so summoning sickness applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a99ce9db-9b7d-45a0-bee7-0875fc62ca48",
		Name:         "Loot, the Nexus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedOneColor(effects.RfCreatureCDifferentPowers),
			Label:          "{T}: Choose a color. Add one mana of that color for each different power among creatures you control",
		}},
	})
}
