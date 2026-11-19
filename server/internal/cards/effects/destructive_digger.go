package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Destructive Digger — Creature — Goblin {2}{R}, 3/2:
//
//	"{3}, {T}, Sacrifice an artifact or land: Draw a card."
func init() {
	Register(Spec{
		OracleID:     "56787f7d-d9e9-42f2-a419-31f15e6e436e",
		Name:         "Destructive Digger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice an artifact or land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{3}"), TapCost(),
				game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact or land", Or(Artifact(), Land()))}),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
