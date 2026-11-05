package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savvy Profiteer — Creature — Ferengi Rogue {1}{R} (Star Trek, #4973), 2/1:
//
//	"{T}: Add {R}.
//	 {3}, {T}: Create a Treasure token. (It's an artifact with "{T},
//	 Sacrifice this token: Add one mana of any color.")"
//
// Both abilities tap the creature, so summoning sickness bars each the
// turn it arrives (CR 302.6). The Treasure is the printed token row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e96a9f9a-9326-4f0b-98ad-23210c97299f",
		Name:         "Savvy Profiteer",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
