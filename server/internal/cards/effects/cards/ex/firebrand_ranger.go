package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firebrand Ranger (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Human Soldier Ranger {1}{R}, 2/1:
//	"{G}, {T}: You may put a basic land card from your hand onto the battlefield."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ceb49f3-a997-468d-9653-8ca5ceb2bcfb",
		Name:         "Firebrand Ranger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: You may put a basic land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Effect: effects.Do(effects.PutFromHandOntoBattlefield{
				Match:    effects.OldB096BasicLandMatch,
				Optional: true,
				Label:    "Firebrand Ranger — you may put a basic land card from your hand onto the battlefield",
			}),
		}},
	})
}
