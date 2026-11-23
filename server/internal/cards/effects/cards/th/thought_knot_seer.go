package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	opponent := func() *game.TargetSpec { return effects.TargetPlayer("target opponent", effects.Opponent()) }
	effects.Register(effects.Spec{
		OracleID:     "cef7396d-2168-4282-bf38-670ca9eebd91",
		Name:         "Thought-Knot Seer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Thought-Knot Seer — target opponent reveals their hand; you choose a nonland card from it and exile it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.TargetRevealsYouChooseExile(effects.Nonland(), "nonland card")(item, effects.NewContext(g, item))
				}), opponent()),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisLeaves("Thought-Knot Seer — target opponent draws a card",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if p := effects.TargetedPlayer(ctx); p != uuid.Nil {
						return effects.DrawCards{Player: p, N: 1}.Apply(ctx)
					}
					return nil
				}))), opponent()), effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})),
		},
	})
}
