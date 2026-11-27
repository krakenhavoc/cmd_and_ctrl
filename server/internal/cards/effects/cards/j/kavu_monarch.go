package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kavu Monarch — Creature — Kavu {2}{R}{R}, 3/3 (Invasion):
//
//	"Kavu creatures have trample.
//	 Whenever another Kavu enters, put a +1/+1 counter on this creature."
//
// A layer 6 trample grant to every Kavu creature at the table (Battering
// Sliver's shape; the Monarch is one itself), and an enters trigger over any
// other Kavu, whoever controls it (CR 603.6a), with the Kavu's creature type
// read as it enters.
//
// No simplification.
func init() {
	const label = "Kavu Monarch — put a +1/+1 counter on this creature"
	effects.Register(effects.Spec{
		OracleID:     "98cc734d-16d5-4c3c-98ca-a45efad27127",
		Name:         "Kavu Monarch",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Kavu"}}, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.HasSubtypeWord("Kavu")
			}, label, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
