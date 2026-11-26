package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kheru Dreadmaw — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Creature — Zombie Crocodile {4}{B}, 4/4: Defender. {1}{G}, Sacrifice another
//	creature: You gain life equal to the sacrificed creature's toughness.
//
// The toughness is the sacrificed creature's as it last existed (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c8c9f16-a3d0-49f2-a95b-dfef756de0b3",
		Name:            "Kheru Dreadmaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice another creature: You gain life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				fed, ok := effects.B17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				return effects.GainLife{Amount: effects.DepartedCreatureToughness(g, fed)}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
