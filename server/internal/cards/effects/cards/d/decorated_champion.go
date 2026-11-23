package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Decorated Champion — Creature — Elf Warrior {1}{G}, 2/2 (Battlebond,
// #5529):
//
//	"Whenever another Warrior your team controls enters, put a +1/+1
//	 counter on this creature."
//
// "Your team" is "you" (CR 102.4). Avatar Enthusiasts' shape with a
// Warrior: any other permanent with the type entering under the
// controller's control (a token, a changeling), read through the layers,
// with "Warrior" a word of the card's text (CR 612.2). The counter goes on
// the Champion if it is still on the battlefield as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7c0052e-6dda-4fcf-9368-7749aa9b9c5c",
		Name:         "Decorated Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Warrior"))
			}, "Decorated Champion — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
