package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rayblade Trooper — Creature — Human Soldier {2}{W}, 2/2 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, put a +1/+1 counter on target creature you
//	 control.
//	 Whenever a nontoken creature you control with a +1/+1 counter on it
//	 dies, create a 1/1 white Human Soldier creature token.
//	 Warp {1}{W}"
//
// The enters trigger targets a creature you control (the Trooper itself
// included). The death watch is Meltstrider Eulogist's, which reads the
// counters the creature had as it last existed on the battlefield
// (CR 603.10a), narrowed to a nontoken creature (Liesa's reading), the
// Trooper itself included. Several such creatures dying together each
// trigger it. The token is the footprinted CreateToken primitive. A warped
// Trooper is exiled at the end step, not dying, and makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "792da586-a687-428c-a10e-75431249df5b",
		Name:             "Rayblade Trooper",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Rayblade Trooper — put a +1/+1 counter on target creature you control",
				effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if !effects.ACreatureYouControlDied(ev, source, ch, g) {
					return false
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Rayblade Trooper — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
