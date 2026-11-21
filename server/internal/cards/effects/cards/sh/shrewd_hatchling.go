package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrewd Hatchling — Creature — Elemental {3}{U/R}, 6/6 (Eventide, #5590):
//
//	"This creature enters with four -1/-1 counters on it.
//	 {U/R}: Target creature can't block this creature this turn.
//	 Whenever you cast a blue spell, remove a -1/-1 counter from this
//	 creature.
//	 Whenever you cast a red spell, remove a -1/-1 counter from this
//	 creature."
//
// The four counters are an entry replacement (Moonshadow's shape), and each
// colour is its own trigger: a spell of both colours removes two. A
// creature with no -1/-1 counter left removes nothing.
// The {U/R} row is Burning-Tree Bloodscale's "can't block this creature"
// record, which bars only blocks of the Hatchling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "011371f0-a7f1-47a5-9883-22a14cfe4de9",
		Name:         "Shrewd Hatchling",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EveB1HatchlingCounters("Shrewd Hatchling")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U/R}: Target creature can't block this creature this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{U/R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBlockThisCreature("Shrewd Hatchling — can't block it"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.EveB1RemoveCounterOnCast("Shrewd Hatchling", "U"),
			effects.EveB1RemoveCounterOnCast("Shrewd Hatchling", "R"),
		},
	})
}
