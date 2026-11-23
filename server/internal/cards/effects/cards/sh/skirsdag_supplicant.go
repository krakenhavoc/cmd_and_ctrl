package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skirsdag Supplicant — Creature — Human Cleric {2}{B}, 2/3 (Eldritch
// Moon, #5541):
//
//	"{B}, {T}, Discard a card: Each player loses 2 life."
//
// The discard is a cost, picked from hand at announce (CR 602.2b); the
// loss is Crushing Disappointment's eachPlayerLosesTwoLife, the
// activator included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7d97311-6e34-4679-a678-e107022c785c",
		Name:         "Skirsdag Supplicant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Discard a card: Each player loses 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue}, // a life loss answers nothing (Orzhov Guildmage)
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EachPlayerLosesTwoLife(effects.NewContext(g, item))
			},
		}},
	})
}
