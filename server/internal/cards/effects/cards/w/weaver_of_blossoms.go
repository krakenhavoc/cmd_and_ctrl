package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weaver of Blossoms // Blossom-Clad Werewolf — {2}{G} Creature — Human
// Werewolf 2/3 // Creature — Werewolf 3/4 (#2586, ADR 0132):
//
//	Front: "{T}: Add one mana of any color.
//	        Daybound"
//	Back:  "{T}: Add two mana of any one color.
//	        Nightbound"
//
// Birds of Paradise's pick, then Gilded Lotus's single colour choice for
// two mana. Summoning sickness applies to both (CR 302.6, enforced by
// the mana-ability path).
//
// No simplification.
func init() {
	const oracle = "a64ecae7-0b09-48e1-8108-89442547ffda"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Weaver of Blossoms",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Blossom-Clad Werewolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "Add two mana of any one color",
		}},
	})
}
