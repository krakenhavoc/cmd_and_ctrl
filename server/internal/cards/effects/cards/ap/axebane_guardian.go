package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Axebane Guardian — Creature — Human Druid {2}{G}, 0/3 (Tarkir:
// Dragonstorm Commander):
//
//	"Defender
//	 {T}: Add X mana in any combination of colors, where X is the number
//	 of creatures you control with defender."
//
// A mana ability (it never uses the stack) in Selvala's shape: X
// independent any-colour picks, counted as the ability is activated and
// the Guardian included. The count reads defender through the layers, so
// a creature that lost it or was granted it counts as it stands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "060e0378-e335-4b3d-82aa-ae36639339f6",
		Name:            "Axebane Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedAnyCombinationOfColors(effects.CmdrB19DefenderProduced),
			Label:          "{T}: Add X mana in any combination of colors, where X is the number of creatures you control with defender",
		}},
	})
}
