package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metathran Aerostat (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Metathran {2}{U}{U}, 2/2:
//	"Flying
//	{X}{U}: You may put a creature card with mana value X from your hand onto the
//	 battlefield. If you do, return this creature to its owner's hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "093daf01-d83b-4d49-a6b3-ec9a799b9582",
		Name:            "Metathran Aerostat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		XMatters:        true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{U}: You may put a creature card with mana value X from your hand onto the battlefield. If you do, return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{X}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.S82OldB645AerostatEffect(g, item, ctx.X())
			},
		}},
	})
}
