package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexavus — Artifact Creature — Construct {6}, 0/0 (The Brothers' War):
//
//	"Flying
//	 This creature enters with six +1/+1 counters on it.
//	 {1}, Remove a +1/+1 counter from this creature: Put a flying counter
//	 on another target creature.
//	 {1}, Remove a counter from another creature you control: Put a +1/+1
//	 counter on this creature."
//
// The entry counters are the printed CR 614.1c clause (Afiya Grove's
// b10EntersWithCounters), so the 0/0 never exists without them. The first
// ability pays a +1/+1 counter off Hexavus (RemoveCountersFromThis) and
// targets another creature, any controller; the keyword counter is an
// ordinary placement (CR 122.1b). The second removes a counter of any kind
// from a creature its controller controls other than Hexavus (Xavier Sal's
// any-kind cost, chosen as it is paid) and puts a +1/+1 counter on Hexavus.
//
// No simplification.
func init() {
	removeCounter := effects.RemoveCountersFrom("", 1, "another creature you control", effects.Creature())
	removeCounter.RemoveCounters.From = effects.Another(removeCounter.RemoveCounters.From)
	effects.Register(effects.Spec{
		OracleID:        "ba48a1a6-a53f-4e7e-881a-472637a3a007",
		Name:            "Hexavus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 6, "Hexavus: enters with six +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, Remove a +1/+1 counter from this creature: Put a flying counter on another target creature.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  effects.PutAFlyingCounterOnTheTarget,
			},
			{
				Label:   "{1}, Remove a counter from another creature you control: Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{1}"), removeCounter),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
