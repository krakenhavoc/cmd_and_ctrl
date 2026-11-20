package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5391f601-b778-413f-a31f-9f8b23b40f5c",
		Name:         "Overbeing of Myth",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			powerIsCardsInYourHand(),
			cmdr2B08ToughnessIsCardsInYourHand(),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginDrawStep, ByYou, "Overbeing of Myth — draw an additional card", Do(DrawCards{N: 1})),
		},
	})
}
