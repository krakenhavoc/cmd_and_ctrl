package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Juju Bubble — Artifact {1} (Visions, S82 older sets batch 789, #5631):
//
//	"Cumulative upkeep {1}
//	 When you play a card, sacrifice this artifact.
//	 {2}: You gain 1 life."
//
// "Play a card" is a land played or a spell cast, never a land put onto the
// battlefield by an effect (Flubs, the Fool's trigger). The sacrifice does
// nothing if the Bubble has already left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9344ae2e-cb2e-400d-9157-e8158ce31813",
		Name:         "Juju Bubble",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Juju Bubble — cumulative upkeep {1}", "{1}"),
			effects.WheneverYouPlayACard("Juju Bubble — sacrifice this artifact", func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return g.SacrificePermanentForEffect(item.SourceCardID)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
