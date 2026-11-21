package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explorer's Cache — Artifact {1}{G} (The Lost Caverns of Ixalan, #4470):
//
//	"This artifact enters with two +1/+1 counters on it.
//	 Whenever a creature you control with a +1/+1 counter on it dies, put
//	 a +1/+1 counter on this artifact.
//	 {T}: Move a +1/+1 counter from this artifact onto target creature.
//	 Activate only as a sorcery."
//
// Costume Closet with two differences: the move may target any creature,
// and the trigger is Meltstrider Eulogist's death watch
// (ACreatureYouControlDied) narrowed by the +1/+1 counters the creature
// had as it last existed on the battlefield (CR 603.10a). The entry
// counters are a CR 614 replacement, so they are there as it enters. The
// move is Simic Fluxmage's: the counter lands on the target first and
// leaves the Cache only if it was put on (CR 122.5). A Cache that has
// left the battlefield puts no counter on anything (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fa0c9d1-2a63-41a8-8733-39bfaee4eb0a",
		Name:         "Explorer's Cache",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Explorer's Cache: enters with two +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Move a +1/+1 counter from this artifact onto target creature. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.SimicFluxmageMove,
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ACreatureYouControlDied(ev, source, ch, g) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Explorer's Cache — put a +1/+1 counter on this artifact", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
