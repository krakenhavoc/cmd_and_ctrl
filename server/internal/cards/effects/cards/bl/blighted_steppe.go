package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blighted Steppe — Land (Battle for Zendikar):
//
//	"{T}: Add {C}.
//	 {3}{W}, {T}, Sacrifice this land: You gain 2 life for each creature you
//	 control."
//
// The creatures are counted as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db16a2fb-dc42-4086-9928-52076043097f",
		Name:         "Blighted Steppe",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, {T}, Sacrifice this land: You gain 2 life for each creature you control.",
			Purpose: game.Purpose{LifeGain: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GainLife{Amount: 2 * effects.CountYouControl(ctx, effects.Creature())}.Apply(ctx)
			},
		}},
	})
}
