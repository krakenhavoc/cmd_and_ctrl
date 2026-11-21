package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "12c57ebe-35a7-4cb3-936b-99c148a3cfed",
		Name:         "Chakra Meditation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Chakra Meditation — return up to one target instant or sorcery card from your graveyard to your hand",
					effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())).WithCount(0, 1)),
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Chakra Meditation — draw a card, then discard a card unless there are three or more Lesson cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: effects.AvB28DrawThenDiscardUnlessThreeLessons, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				}),
		},
	})
}
