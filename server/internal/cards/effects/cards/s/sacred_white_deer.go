package cardss

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sacred White Deer (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	{3}{G}, {T}: You gain 4 life. Activate only if you control a Yanggu planeswalker.
//
// {1}{G} Creature — Elk 2/2. The activation condition (CR 602.1b) is a Yanggu
// planeswalker, by subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d7bef31-3447-4de6-9d0e-bd6749e7ecd5",
		Name:         "Sacred White Deer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}: You gain 4 life. Activate only if you control a Yanggu planeswalker.",
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost()),
			Purpose: game.Purpose{LifeGain: 4, Answers: game.AnswerValue},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.B42ControlsPlaneswalkerNamed(g, controller, "Yanggu")
			},
			Effect: effects.Do(effects.GainLife{Amount: 4}),
		}},
	})
}
