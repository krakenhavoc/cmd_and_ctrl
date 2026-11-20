package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malakir Cullblade — Creature — Vampire Warrior {1}{B}, 1/1 (Magic Origins, #5549):
//
//	"Whenever a creature an opponent controls dies, put a +1/+1 counter on this creature."
//
// Yahenni's counter trigger: a creature that died under an opponent's
// control, read as it last existed (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fbaf052-a8c3-4395-8f40-0cb3a10d5ccd",
		Name:         "Malakir Cullblade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) != source.Controller
			}, "Malakir Cullblade — put a +1/+1 counter on Malakir Cullblade",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
