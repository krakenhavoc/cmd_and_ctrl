package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borrowed Knowledge — Sorcery {2}{R}{W}, Secrets of Strixhaven (#4456):
//
//	"Choose one —
//	 • Discard your hand, then draw cards equal to the number of cards in
//	   target opponent's hand.
//	 • Discard your hand, then draw cards equal to the number of cards
//	   discarded this way."
//
// "Discard your hand" takes every card, so there is no choice to make and the
// random-order discard primitive is exactly the printed instruction (Forgotten
// Creation, Change of Fortune). Both draws wait for the discard to settle (a
// replacement offer such as madness, CR 608.2c) as the tail of the keyed
// discard. The first bullet counts the opponent's hand after the discard
// (CR 608.2c, "then"), and does nothing when its target is gone (CR 608.2b).
// The second draws the cards that REALLY reached a graveyard, the shared
// forgottenCreationDraw tail ("that many").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d6507b0-c986-49d1-aea3-77b2cf2a7433",
		Name:         "Borrowed Knowledge",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Discard your hand, then draw cards equal to the number of cards in target opponent's hand.",
				TargetPlayer("target opponent", Opponent()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					if t, ok := ModeTarget(ctx, occ); !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return socB21DiscardHandThen(ctx, item, borrowedKnowledgeOpponentsHand)
				}),
			ModeDoing("Discard your hand, then draw cards equal to the number of cards discarded this way.",
				nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return socB21DiscardHandThen(ctx, item, forgottenCreationDraw)
				}),
		),
	})
}
