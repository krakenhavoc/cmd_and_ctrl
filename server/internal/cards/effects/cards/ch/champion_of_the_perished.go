package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Champion of the Perished — Creature — Zombie {B}, 1/1 (EDHREC rank
// 2777):
//
//	"Whenever another Zombie you control enters, put a +1/+1 counter
//	 on this creature."
//
// Champion of the Parish for Zombies. "Another Zombie you control" is
// any other permanent with the type entering under the controller's
// control — a token, a changeling, a Zombie stolen and re-entering —
// read post-layer. The counter goes on the Champion if it is still
// on the battlefield when the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3654fbd-16a5-4953-84ac-534e8421032f",
		Name:         "Champion of the Perished",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Zombie"))
			}, "Champion of the Perished — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
