package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0c4e2c90-c17b-42cc-b4d7-cf75970fbe90",
		Name:         "Staff of the Storyteller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Staff of the Storyteller — create a 1/1 white Spirit creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(NewContext(g, item))
				}),
			OncePerBatch(On(game.EventTokenCreated, rfReprintBYouCreatedACreatureToken,
				"Staff of the Storyteller — put a story counter on it", putACounterOnThis(story))),
		},
		Activated: []ActivatedAbility{{
			Label: "{W}, {T}, Remove a story counter from this artifact: Draw a card.",
			Cost:  Plus(ManaCost("{W}"), TapCost(), RemoveCountersFromThis(story, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
