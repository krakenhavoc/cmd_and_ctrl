package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sokka, Bold Boomeranger — Legendary Creature — Human Warrior Ally
// {U}{R}, 1/1:
//
//	"When Sokka enters, discard up to two cards, then draw that many
//	 cards.
//	 Whenever you cast an artifact or Lesson spell, put a +1/+1 counter
//	 on Sokka."
//
// The enters trigger is Tersa Lightshatter's (b39MayDiscardThenDraw):
// choose zero to two cards, discard them, draw as many as were
// discarded. The cast trigger is Dr. Madison Li's shape over "artifact
// or Lesson", on the counter-on-this primitive.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "bdbec6d5-0f10-4114-b709-4402a44debdf",
		Name:         "Sokka, Bold Boomeranger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sokka, Bold Boomeranger — discard up to two cards, then draw that many",
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(2, false,
						"Sokka, Bold Boomeranger — discard up to two cards, then draw that many",
						drawAsManyAsDiscarded)(NewContext(g, item))
				}),
			WheneverYouCast(Or(Artifact(), HasSubtype("Lesson")),
				"Sokka, Bold Boomeranger — put a +1/+1 counter on Sokka",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
