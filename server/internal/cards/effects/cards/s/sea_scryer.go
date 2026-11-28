package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Scryer — Creature — Merfolk Wizard {1}{U}, 1/1 (Mirage, #5633):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {U}."
//
// Two mana abilities (CR 605.1a), each tapping the creature; the second
// pays {1} first, so a lone Scryer nets nothing and needs another source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f17954dd-c8e0-4983-b81e-5440cca41afc",
		Name:         "Sea Scryer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{U}",
				Label:    "{1}, {T}: Add {U}.",
			},
		},
	})
}
