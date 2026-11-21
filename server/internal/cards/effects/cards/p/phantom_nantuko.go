package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantom Nantuko — Creature — Insect Spirit {2}{G}, 0/0:
//
//	"Trample
//	 This creature enters with two +1/+1 counters on it.
//	 If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.
//	 {T}: Put a +1/+1 counter on this creature."
//
// ADR 0108 §8 (#1906): one of the Phantoms (phantoms.go). The tap
// ability puts its counter on the creature only while it is still the
// object that activated it (putCounterOnSelf).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0951b529-646c-4dfd-88ad-84ee117ce722",
		Name:            "Phantom Nantuko",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    effects.PhantomReplacements("Phantom Nantuko", 2),
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Effect:  effects.PutCounterOnSelf,
		}},
	})
}
