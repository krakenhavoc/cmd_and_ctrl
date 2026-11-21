package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temur Sabertooth — Creature — Cat {2}{G}{G}, 4/3 (Khans of Tarkir):
//
//	"{1}{G}: You may return another creature you control to its owner's
//	 hand. If you do, this creature gains indestructible until end of turn."
//
// Nothing targets, so the creature is a resolution-time pick (CR 608.2),
// zero being the "you may" decline, and the Sabertooth itself is never
// offered (First Responder's and Loyal Gryff's shape). "If you do" is told
// whether the creature really reached a hand (CR 400.7): a commander that
// went to the command zone instead gives nothing. The grant goes only to
// this same permanent still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bf8e846-9d10-4dbc-a379-df37bb9fb6d4",
		Name:         "Temur Sabertooth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{G}: You may return another creature you control to its owner's hand. If you do, this creature gains indestructible until end of turn.",
			Cost:  effects.ManaCost("{1}{G}"),
			// The indestructible half answers a destroy spell or lethal damage.
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ChooseAnotherCreatureYouControl(g, item,
					"Temur Sabertooth — return another creature you control to its owner's hand?", effects.RprB23SabertoothReturn)
			},
		}},
	})
}
