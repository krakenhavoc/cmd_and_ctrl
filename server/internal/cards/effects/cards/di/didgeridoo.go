package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Didgeridoo — Artifact {1}, (Homelands, S82 OLD b821, #5635):
//
//	"{3}: You may put a Minotaur permanent card from your hand onto the
//	battlefield."
//
// Goblin Lackey's put from hand (PutFromHandOntoBattlefield, a "you may"
// answered as the ability resolves) behind a {3} cost.
//
// No simplification.
func init() {
	const label = "Didgeridoo — you may put a Minotaur permanent card from your hand onto the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "e9a2cb10-6526-4a5c-bda7-e2f328b377dc",
		Name:         "Didgeridoo",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: You may put a Minotaur permanent card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.Subtype("Minotaur"),
					Optional: true,
					Label:    label,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
