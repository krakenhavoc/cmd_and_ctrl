package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unclaimed Territory — Land (EDHREC rank 247):
//
//	"As this land enters, choose a creature type.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell of the chosen type."
//
// Cavern of Souls minus the "can't be countered" clause: the same CR
// 614.12 chosen-type prompt (ChooseCreatureTypeAsEnters) and the same
// spend restriction (ChosenTypeManaRestrictions), both built for
// exactly this shape and reused verbatim.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "584b15f2-6ae9-413a-8b8d-9244dbea4878",
		Name:         "Unclaimed Territory",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Unclaimed Territory"),
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:             effects.ManaAbilityCost{Tap: true},
				Produced:         "{W|U|B|R|G}",
				Label:            "Add one mana of any color (chosen-type creature spells only)",
				RestrictionsFunc: effects.ChosenTypeManaRestrictions(),
			},
		},
	})
}
