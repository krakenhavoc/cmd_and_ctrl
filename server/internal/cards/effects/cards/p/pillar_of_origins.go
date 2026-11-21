package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillar of Origins — Artifact {2} (Commander Masters):
//
//	"As this artifact enters, choose a creature type.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell of the chosen type."
//
// Unclaimed Territory's two halves on an artifact: the CR 614.12
// chosen-type prompt (ChooseCreatureTypeAsEnters) and the spend
// restriction read from the chosen type (ChosenTypeManaRestrictions).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e62f30d-3130-416e-b459-48b5f71f8be9",
		Name:         "Pillar of Origins",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Pillar of Origins"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:             effects.ManaAbilityCost{Tap: true},
			Produced:         "{W|U|B|R|G}",
			Label:            "Add one mana of any color (chosen-type creature spells only)",
			RestrictionsFunc: effects.ChosenTypeManaRestrictions(),
		}},
	})
}
