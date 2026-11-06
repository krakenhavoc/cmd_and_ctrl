package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Enthusiasts — Creature — Human Peasant Ally {2}{W}, 2/2 (Avatar:
// The Last Airbender, #4459):
//
//	"Whenever another Ally you control enters, put a +1/+1 counter on
//	 this creature."
//
// Champion of the Perished's shape for Allies: any other permanent with
// the type entering under the controller's control (a token, a
// changeling), read post-layer, with the type a word of the card's text
// (CR 612.2). The counter goes on the Enthusiasts if they are still on
// the battlefield when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b0d9d79-2e86-4079-8db3-500483c8807e",
		Name:         "Avatar Enthusiasts",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
			}, "Avatar Enthusiasts — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
