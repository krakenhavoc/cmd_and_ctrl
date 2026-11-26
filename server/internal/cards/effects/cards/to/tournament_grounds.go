package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tournament Grounds — Land (Throne of Eldraine):
//
//	"{T}: Add {C}.
//	 {T}: Add {R}, {W}, or {B}. Spend this mana only to cast a Knight or
//	 Equipment spell."
//
// Fili and Kili's shape: one CR 106.6 cast restriction whose subtype tag
// lists both alternatives, read on the spell's effective subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3ed5d7f-c554-4d46-86e5-7bb8df2693f6",
		Name:         "Tournament Grounds",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{R|W|B}",
				Label:        "Add {R}, {W}, or {B}. Spend this mana only to cast a Knight or Equipment spell",
				Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictSubtype("Knight|Equipment")},
			},
		},
	})
}
