package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tributary Instructor — Creature — Merfolk Shaman {3}{G}, 4/4 (Lost
// Caverns of Ixalan Commander, #5368):
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 Whenever a creature you control with a +1/+1 counter on it dies, draw a
//	 card."
//
// Mentor is the shared trigger (Blade Instructor). The death watch is
// Meltstrider Eulogist's: counters are read as the creature last existed
// on the battlefield (CR 603.10a), the Instructor itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c70b9d8-b04e-48f8-a4e0-12ce077d22f2",
		Name:         "Tributary Instructor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Mentor("Tributary Instructor — mentor"),
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ACreatureYouControlDied(ev, source, ch, g) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Tributary Instructor — draw a card", effects.Do(effects.DrawCards{N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
