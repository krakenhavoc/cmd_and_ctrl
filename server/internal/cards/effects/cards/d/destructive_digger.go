package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Destructive Digger — Creature — Goblin {2}{R}, 3/2:
//
//	"{3}, {T}, Sacrifice an artifact or land: Draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56787f7d-d9e9-42f2-a419-31f15e6e436e",
		Name:         "Destructive Digger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice an artifact or land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{3}"), effects.TapCost(),
				game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an artifact or land", effects.Or(effects.Artifact(), effects.Land()))}),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
