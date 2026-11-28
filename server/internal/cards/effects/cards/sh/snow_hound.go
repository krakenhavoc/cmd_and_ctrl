package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snow Hound — Creature — Dog {2}{W}, 1/1 (Ice Age, S82 OLD #5636):
//
//	"{1}, {T}: Return this creature and target green or blue creature you control
//	 to their owner's hand."
//
// Wizard Mentor's two-body bounce, with a colour-restricted target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fe8512f-826c-400d-ba6d-1de4f315b617",
		Name:         "Snow Hound",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Return this creature and target green or blue creature you control to their owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target green or blue creature you control", effects.YouControl(), effects.Or(effects.OfColor("G"), effects.OfColor("U"))),
			Effect:  effects.BounceThisAndTargetB831,
		}},
	})
}
