package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thought-Knot Seer — Creature — Eldrazi {3}{C}, 4/4 (Special Guests):
//
//	"When this creature enters, target opponent reveals their hand. You
//	 choose a nonland card from it and exile that card.
//	 When this creature leaves the battlefield, target opponent draws a
//	 card."
//
// Two triggers. The enters one is the revealed-hand pick (ADR 0116): the
// whole table sees the hand (CR 701.20a), the exile is permanent (it is not a
// discard, CR 701.9a), and a hand with no nonland card is revealed and
// nothing is exiled (CR 609.3). The leaves one targets when it triggers, so
// it draws for that opponent even though the Seer is gone (CR 603.10a).
//
// No simplification.
func init() {
	opponent := func() *game.TargetSpec { return TargetPlayer("target opponent", Opponent()) }
	Register(Spec{
		OracleID:     "cef7396d-2168-4282-bf38-670ca9eebd91",
		Name:         "Thought-Knot Seer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Thought-Knot Seer — target opponent reveals their hand; you choose a nonland card from it and exile it",
				func(g *game.Game, item *game.StackItem) error {
					return TargetRevealsYouChooseExile(Nonland(), "nonland card")(item, NewContext(g, item))
				}), opponent()),
			TriggerWithPurpose(Targeting(WhenThisLeaves("Thought-Knot Seer — target opponent draws a card",
				Do(Instruction(func(ctx *Context) error {
					if p := TargetedPlayer(ctx); p != uuid.Nil {
						return DrawCards{Player: p, N: 1}.Apply(ctx)
					}
					return nil
				}))), opponent()), ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})),
		},
	})
}
