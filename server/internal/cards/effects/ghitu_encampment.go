package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghitu Encampment — Land (Urza's Legacy, #5623):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {1}{R}: This land becomes a 2/1 red Warrior creature with first
//	 strike until end of turn. It's still a land. (It deals combat
//	 damage before creatures without first strike.)"
//
// A manland built like Faerie Conclave, through the shared
// ulgB1AnimateThisLand row: until end of turn it is a 2/1 red Warrior
// creature with first strike and still a land, with its mana ability. It
// enters tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cc02c16-8cf9-4ac7-9475-231c27121968",
		Name:         "Ghitu Encampment",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{
			ulgB1AnimateThisLand("Ghitu Encampment", "{1}{R}",
				"{1}{R}: This land becomes a 2/1 red Warrior creature with first strike until end of turn. It's still a land. (It deals combat damage before creatures without first strike.)",
				2, 1, "Warrior", "R", "first strike"),
		},
	})
}
