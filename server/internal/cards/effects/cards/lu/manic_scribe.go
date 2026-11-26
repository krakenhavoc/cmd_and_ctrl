package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manic Scribe — Creature — Human Wizard {1}{U}, 0/3 (Shadows over
// Innistrad, #5543):
//
//	"When this creature enters, each opponent mills three cards.
//	 Delirium — At the beginning of each opponent's upkeep, if there are
//	 four or more card types among cards in your graveyard, that player
//	 mills three cards."
//
// The upkeep trigger is an intervening if (CR 603.4) on the Scribe's
// controller's own graveyard, checked as each opponent's upkeep begins and
// again as the trigger resolves; "that player" is the upkeep's player, read
// off the event. A library of fewer than three cards mills what it has
// (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2ef475d-6043-4da8-83f0-dffa76052707",
		Name:         "Manic Scribe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Manic Scribe — each opponent mills three cards", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillEachPlayerInTurn(ctx, ctx.Opponents(), effects.ConstantMill(3), game.KeyedTail{})
			}),
			effects.InterveningIf(effects.B24CondDelirium{},
				effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent, effects.YouHaveDeliriumWhen),
					"Manic Scribe — delirium: that player mills three cards", func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						player := ctx.Trigger().Event.Actor
						if player == uuid.Nil {
							return nil
						}
						return effects.MillCards{Player: player, N: 3}.Apply(ctx)
					})),
		},
	})
}
