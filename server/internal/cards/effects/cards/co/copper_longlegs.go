package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Copper Longlegs — Creature — Phyrexian Spider {1}{G}, 1/3 (Phyrexia: All
// Will Be One, S82):
//
//	"Reach
//	 {1}{G}, Sacrifice this creature: Proliferate."
//
// Contagion Clasp's activated proliferate with a sacrifice in the cost; the
// controller chooses what to proliferate as the ability resolves, after the
// Longlegs has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "066f6a3d-5e7b-40a4-b348-6dfb5d5239de",
		Name:            "Copper Longlegs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this creature: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Proliferate{}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
