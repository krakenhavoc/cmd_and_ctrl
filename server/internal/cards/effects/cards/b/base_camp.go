package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Base Camp — Land (Zendikar Rising, #5508):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Cleric,
//	 Rogue, Warrior, or Wizard spell or to activate an ability of a Cleric,
//	 Rogue, Warrior, or Wizard."
//
// Brotherhood Headquarters' shape: one CR 106.6 restriction tag with eight
// alternatives (a cast of, or an activation of an ability of, each of the
// four party types), the types judged on the effective subtypes.
//
// No simplification.
func init() {
	var alternatives [][]string
	for _, t := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
		alternatives = append(alternatives,
			[]string{effects.ManaRestrictCast, effects.ManaRestrictSubtype(t)},
			[]string{effects.ManaRestrictActivate, effects.ManaRestrictSubtype(t)})
	}
	effects.Register(effects.Spec{
		OracleID:     "41fbf835-baee-4530-9155-e2c1b9045567",
		Name:         "Base Camp",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Label:        "Add one mana of any color. Spend this mana only to cast a Cleric, Rogue, Warrior, or Wizard spell or to activate an ability of a Cleric, Rogue, Warrior, or Wizard",
				Restrictions: []string{game.ManaRestrictAnyOf(alternatives...)},
			},
		},
	})
}
