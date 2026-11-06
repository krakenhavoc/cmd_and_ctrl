package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Accumulate Wisdom — Instant — Lesson {1}{U} (Avatar: The Last
// Airbender, #4459):
//
//	"Look at the top three cards of your library. Put one of those cards
//	 into your hand and the rest on the bottom of your library in any
//	 order. Put each of those cards into your hand instead if there are
//	 three or more Lesson cards in your graveyard."
//
// The Lesson cards are counted as the spell resolves (CR 608.2h), while
// Accumulate Wisdom itself is still on the stack and so not among them
// (Combustion Technique's count). With fewer than three it is Impulse's
// sentence (LookAtTopThenTakeOneRestOnBottom): one card, mandatory, and
// you order the rest on the bottom. With three or more, the same look puts
// every one of the cards into your hand and nothing goes to the bottom.
//
// No simplification.
func init() {
	lessons := g2CardsInYourGraveyard(g2IsLessonCard)
	impulse := LookAtTopThenTakeOneRestOnBottom(3, "Accumulate Wisdom — put one of those cards into your hand")
	Register(Spec{
		OracleID:     "a63eb744-71c4-447d-9372-c3a734f022c5",
		Name:         "Accumulate Wisdom",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if lessons(item, ctx) < 3 {
				return impulse(ctx.Game, item)
			}
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				All:    true,
				Label:  "Accumulate Wisdom — put each of those cards into your hand",
				Key:    takeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
