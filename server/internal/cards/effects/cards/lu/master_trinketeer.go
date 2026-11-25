package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master Trinketeer — Creature — Dwarf Artificer {2}{W}, 3/2 (Kaladesh,
// #5539):
//
//	"Servos and Thopters you control get +1/+1.
//	 {3}{W}: Create a 1/1 colorless Servo artifact creature token."
//
// The anthem is Imperious Perfect's layer 7c static over Servos and Thopters
// the controller controls (effective subtypes, so a changeling counts), and
// the Trinketeer is neither, so it is not pumped. "Servo" and "Thopter" are
// words of the source's text (CR 612.2). The ability is an ordinary activation
// and each Servo it makes is pumped the moment it lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa57f607-c5eb-45a7-bfb1-adce0aef49c7",
		Name:         "Master Trinketeer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Servos and Thopters you control get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() &&
					(target.HasSubtype(source.TextSubtype("Servo")) || target.HasSubtype(source.TextSubtype("Thopter")))
			},
			Apply: effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Create a 1/1 colorless Servo artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{3}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
