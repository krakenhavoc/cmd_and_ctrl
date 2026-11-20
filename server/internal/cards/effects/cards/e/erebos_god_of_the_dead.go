package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erebos, God of the Dead — Legendary Enchantment Creature — God
// {3}{B}, 5/7:
//
//	"Indestructible
//	 As long as your devotion to black is less than five, Erebos isn't
//	 a creature.
//	 Your opponents can't gain life.
//	 {1}{B}, Pay 2 life: Draw a card."
//
// The God clause is the shared Theros static (godUnlessDevotion,
// theros_gods.go). "Your opponents can't gain life" is ADR 0107 §5's
// battlefield static (CR 119.7, #1880); it works whether or not Erebos
// is a creature, because it is an ability of the permanent either way.
// The draw costs a real life payment (CR 119.4), so a controller who
// can't pay 2 life can't activate it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1cfc7b12-a595-492d-81ee-bd100ba7de6a",
		Name:            "Erebos, God of the Dead",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{effects.GodUnlessDevotion("B")},
		CantGainLife:    effects.OpponentsCantGainLife(),
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Pay 2 life: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.PayLife(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
