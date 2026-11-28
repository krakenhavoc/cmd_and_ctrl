package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sliver Queen — Legendary Creature — Sliver {W}{U}{B}{R}{G}, 7/7 (Stronghold,
// S82 OLD batch 746, #5627):
//
//	"{2}: Create a 1/1 colorless Sliver creature token."
//
// Sliver Hive's token ability with no condition: the creation of the shared
// 1/1 colorless Sliver row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8376cca-ea96-478a-8e98-c4482031300a",
		Name:         "Sliver Queen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Create a 1/1 colorless Sliver creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 colorless Sliver"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
