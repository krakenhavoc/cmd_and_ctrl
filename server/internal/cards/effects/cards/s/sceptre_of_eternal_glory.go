package cardss

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Sceptre of Eternal Glory — Legendary Artifact {4} (Warhammer 40,000
// Commander, S82 GAP batch 186, #5948):
//
//	"{T}: Add one mana of any color.
//	 {T}: Add three mana of any one color. Activate only if you control
//	 three or more lands with the same name."
//
// Two mana abilities. The second's restriction is Endless Atlas's: three
// lands you control share a name, judged as it is activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce08aff3-efe6-4186-9557-274e00a7ca38",
		Name:         "Sceptre of Eternal Glory",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "{T}: Add one mana of any color."},
			{
				Cost:      effects.ManaAbilityCost{Tap: true},
				Produced:  effects.OneColorOfAmount(3),
				Label:     "{T}: Add three mana of any one color. Activate only if you control three or more lands with the same name.",
				Condition: effects.GapB186ThreeLandsSharingAName,
			},
		},
	})
}
