package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forbidding Watchtower — Land (Urza's Legacy, #5623):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 {1}{W}: This land becomes a 1/5 white Soldier creature until end
//	 of turn. It's still a land."
//
// A manland built like Faerie Conclave, through the shared
// ulgB1AnimateThisLand row: until end of turn it is a 1/5 white Soldier
// creature and still a land, with its mana ability. It enters tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cabf7953-0fac-4dbb-b3ae-05e85e02b3fc",
		Name:         "Forbidding Watchtower",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{
			effects.UlgB1AnimateThisLand("Forbidding Watchtower", "{1}{W}",
				"{1}{W}: This land becomes a 1/5 white Soldier creature until end of turn. It's still a land.",
				1, 5, "Soldier", "W"),
		},
	})
}
