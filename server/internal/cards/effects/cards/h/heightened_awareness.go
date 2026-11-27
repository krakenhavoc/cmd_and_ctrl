package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heightened Awareness — Enchantment {3}{U}{U} (Prophecy, S82 OLD batch 675,
// #5617):
//
//	"As this enchantment enters, discard your hand.
//	 At the beginning of your draw step, draw an additional card."
//
// The first ability is an "as enters" effect (CR 614.12, Spec.AsEnters), not a
// trigger: the controller discards their whole hand (Bedlam Reveler's
// DiscardWholeHand) as it arrives, before anything can respond. The second is
// Midnight Oil's draw-step trigger: one more card at the start of its
// controller's draw step, on top of the turn's own draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a69a0720-35fb-450d-8d48-8e51a14ffa49",
		Name:         "Heightened Awareness",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			_, err := effects.DiscardWholeHand(ctx.Game, card.Controller)
			return err
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginDrawStep, effects.ByYou, "Heightened Awareness — draw an additional card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
