package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of the Storyteller — Artifact {1}{W}:
//
//	"When this artifact enters, create a 1/1 white Spirit creature
//	 token with flying.
//	 Whenever you create one or more creature tokens, put a story
//	 counter on this artifact.
//	 {W}, {T}, Remove a story counter from this artifact: Draw a card."
//
// The counter trigger is batched ("one or more"): it fires on the
// first creature token of a batch of creations and declines the rest.
// The Spirit this artifact makes on entering also counts, so it
// enters with that trigger already waiting. The counter cost is paid
// at announce, so a response can't spend the same counter twice.
//
// No simplification.
func init() {
	const story = "story"
	effects.Register(effects.Spec{
		OracleID:     "0c4e2c90-c17b-42cc-b4d7-cf75970fbe90",
		Name:         "Staff of the Storyteller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Staff of the Storyteller — create a 1/1 white Spirit creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.OncePerBatch(effects.On(game.EventTokenCreated, effects.RfReprintBYouCreatedACreatureToken,
				"Staff of the Storyteller — put a story counter on it", effects.PutACounterOnThis(story))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Remove a story counter from this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.RemoveCountersFromThis(story, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
