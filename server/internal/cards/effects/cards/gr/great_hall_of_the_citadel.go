package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Hall of the Citadel — Land (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add two mana in any combination of colors. Spend this mana only
//	 to cast legendary spells."
//
// Two mana abilities. The second is Cascading Cataracts' paid
// any-combination ability (the {1} is paid with the tap, so it nets one mana),
// with the spend restriction Rockface Village puts on its creature mana: only
// to cast a spell (not to activate an ability), and only a legendary one
// (CR 106.6, the supertype read from the spell being cast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3e28bcf-0ed0-4406-b615-68ddc55b349a",
		Name:         "Great Hall of the Citadel",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced:     effects.AnyCombinationOfColors(2),
				Label:        "{1}, {T}: Add two mana in any combination of colors. Spend this mana only to cast legendary spells.",
				Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictSupertype("Legendary")},
			},
		},
	})
}
