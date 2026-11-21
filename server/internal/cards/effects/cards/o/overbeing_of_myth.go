package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overbeing of Myth — Creature — Spirit Avatar {G/U}{G/U}{G/U}{G/U}{G/U}, */* (Eventide, #5590):
//
//	"Overbeing of Myth's power and toughness are each equal to the number of
//	 cards in your hand.
//	 At the beginning of your draw step, draw an additional card."
//
// Body of Knowledge's characteristic-defining ability (CR 604.3, layer 7a)
// and The Immortal Sun's draw-step trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5391f601-b778-413f-a31f-9f8b23b40f5c",
		Name:         "Overbeing of Myth",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PowerIsCardsInYourHand(),
			effects.Cmdr2B08ToughnessIsCardsInYourHand(),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginDrawStep, effects.ByYou, "Overbeing of Myth — draw an additional card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
