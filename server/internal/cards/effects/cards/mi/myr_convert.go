package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Convert — Artifact Creature — Phyrexian Myr {2}, 2/1
// (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 1
//	 {T}, Pay 2 life: Add one mana of any color."
//
// Toxic is printed, so it is read through the layered view like any
// keyword. The ability is a mana ability (CR 605.1a) whose cost taps the
// Convert and pays 2 life; the life is checked before anything is paid
// (CR 119.4), so at 1 life it can't be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6abf433e-27da-4abb-b262-3af792a1db23",
		Name:            "Myr Convert",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Life: 2},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Pay 2 life: Add one mana of any color.",
		}},
	})
}
