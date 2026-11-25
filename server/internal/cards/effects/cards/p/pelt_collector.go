package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pelt Collector — Creature — Elf Warrior {G}, 1/1 (Guilds of Ravnica):
//
//	"Whenever another creature you control enters or dies, if that
//	 creature's power is greater than this creature's, put a +1/+1 counter on
//	 this creature.
//	 As long as this creature has three or more +1/+1 counters on it, it has
//	 trample."
//
// An intervening "if" (CR 603.4), Sharp-Eyed Rookie's shape: it is read as the
// creature enters or dies and again as the ability resolves, the other
// creature as it is now or as it last existed once it has left (CR 608.2h)
// and this creature's power as it is now or as it last existed. The trample is
// a layer-6 grant to itself while it has three +1/+1 counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2348f3d4-02cc-4611-a7e0-b77604d5f2e8",
		Name:         "Pelt Collector",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters[game.CounterPlusOne] >= 3
			}, "trample"), "As long as this creature has three or more +1/+1 counters on it, it has trample."),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(peltCollectorOutpowers{}, effects.OnAny([]game.EventKind{game.EventETB, game.EventLTB},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B16AnotherCreatureYouControlEnteredOrDied(ev, source, g)
				},
				"Pelt Collector — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

// peltCollectorOutpowers is "that creature's power is greater than this
// creature's", both last-known once they have left (CR 608.2h).
type peltCollectorOutpowers struct{}

func (peltCollectorOutpowers) Holds(ctx *effects.Context) bool {
	other, ok := ctx.TriggeringPermanent()
	if !ok {
		return false
	}
	src, ok := ctx.SourcePermanent()
	if !ok {
		return false
	}
	return other.Power > src.Power
}

func (peltCollectorOutpowers) Condition() (game.FootprintStep, bool) {
	return effects.CondStep(game.FootprintCondBoard, 0), true
}
