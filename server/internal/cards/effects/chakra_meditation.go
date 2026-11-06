package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chakra Meditation — Enchantment {2}{U} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"When this enchantment enters, return up to one target instant or
//	 sorcery card from your graveyard to your hand.
//	 Whenever you cast an instant or sorcery spell, draw a card. Then
//	 discard a card unless there are three or more Lesson cards in your
//	 graveyard."
//
// The cast trigger resolves above the spell that caused it, so a Lesson
// being cast is still on the stack and not counted, as the ruling has
// it. The Lessons are counted after the draw, as the discard would be
// made (avB28DrawThenDiscardUnlessThreeLessons).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12c57ebe-35a7-4cb3-936b-99c148a3cfed",
		Name:         "Chakra Meditation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Chakra Meditation — return up to one target instant or sorcery card from your graveyard to your hand",
					returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())).WithCount(0, 1)),
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Chakra Meditation — draw a card, then discard a card unless there are three or more Lesson cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: avB28DrawThenDiscardUnlessThreeLessons, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				}),
		},
	})
}
