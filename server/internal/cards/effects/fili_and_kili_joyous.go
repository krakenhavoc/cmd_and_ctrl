package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fíli and Kíli, Joyous — Legendary Creature — Dwarf Bard {2}{R}, 3/3
// (The Hobbit):
//
//	"Haste
//	 {T}: Add {R}{R}. Spend this mana only to cast Dwarf, Equipment, and Saga
//	 spells."
//
// Haste is a printed keyword, so the tap ability is usable the turn the pair
// arrives. The mana is fixed red, restricted to casting (not activating) a
// spell with any of the three subtypes. The subtype tag is one `|`-joined
// alternation (ManaRestrictSubtype's documented OR), because tags AND: it
// matches the spell's effective subtypes, so a Dwarf creature, an Equipment
// artifact or a Saga enchantment is payable, while a spell that is none of
// them, or an ability, is refused.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "392f6cd3-3895-42b3-8bb9-6b0c6a689a29",
		Name:            "Fíli and Kíli, Joyous",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{R}{R}",
			Label:        "Add {R}{R}. Spend this mana only to cast Dwarf, Equipment, and Saga spells.",
			Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Dwarf|Equipment|Saga")},
		}},
	})
}
