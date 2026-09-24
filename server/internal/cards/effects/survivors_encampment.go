package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Survivors' Encampment — Land — Desert:
//
//	{T}: Add {C}.
//	{T}, Tap an untapped creature you control: Add one mana of any color.
//
// #758's fixed-count TapOthers component is the second cost. It is
// deliberately not the {T} symbol: a creature that entered this turn
// may pay it, while the Encampment itself pays its own tap symbol.
func init() {
	Register(Spec{
		OracleID:     "152e7e91-4eda-4e72-a9fb-bd5cb2e68239",
		Name:         "Survivors' Encampment",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: ManaAbilityCost{
					Tap: true,
					TapOthers: &game.TapOthersCost{
						Count:  1,
						Filter: TargetPermanent("an untapped creature you control", Creature()),
						Label:  "an untapped creature you control",
					},
				},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
		},
	})
}
