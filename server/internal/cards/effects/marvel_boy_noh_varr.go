package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marvel Boy, Noh-Varr — Legendary Creature — Kree Soldier Hero {1}{R/W},
// 1/1 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever another creature you control enters and whenever you activate
//	 a power-up ability, put a +1/+1 counter on Marvel Boy."
//
// Two events, one counter each: another creature of yours entering
// (Hulkling's enteredUnderYourControl), and an activation by you of a
// power-up ability (APowerUpAbility reads the bit the announcement stamped,
// CR 602.2, so the counter does not wait for the ability to resolve). The
// activating player must be Marvel Boy's controller; a power-up ability of
// an opponent's creature does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a315db3-e66c-4648-b7a0-5aef2ce9b580",
		Name:         "Marvel Boy, Noh-Varr",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventActivateAbility},
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventActivateAbility {
						return ev.Actor == source.Controller && APowerUpAbility(ev, source, ch, g)
					}
					c, ok := enteredUnderYourControl(ev, source, g, true)
					return ok && c.IsCreature()
				},
				"Marvel Boy — put a +1/+1 counter on Marvel Boy",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
