package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tainted Isle — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add {U} or {B}. Activate only if you control a Swamp."
//
// Tainted Peak's shape in blue and black: ManaAbility.Condition gates
// the coloured ability on the EFFECTIVE land subtypes, so an Urborg-ed
// board enables it and a Blood Moon one does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0222414f-98b5-458a-a0fd-831a66cd8b07",
		Name:         "Tainted Isle",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:      effects.ManaAbilityCost{Tap: true},
				Produced:  "{U|B}",
				Label:     "{T}: Add {U} or {B}. Activate only if you control a Swamp",
				Condition: effects.ControlsAtLeast(1, effects.MatchLandSubtype("Swamp")),
			},
		},
	})
}
