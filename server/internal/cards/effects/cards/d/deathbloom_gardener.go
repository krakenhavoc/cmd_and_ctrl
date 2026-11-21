package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathbloom Gardener — Creature — Elf Druid {2}{G}, 1/1 (Dominaria United):
//
//	"Deathtouch
//	 {T}: Add one mana of any color."
//
// Birds of Paradise's mana ability (a colour pipe, CR 106.1b) with the
// deathtouch keyword the combat engine reads. Summoning sickness applies
// to the tap (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c7f9701-c39e-4f14-844a-15f2e3c41ca9",
		Name:            "Deathbloom Gardener",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
