package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenger Folk — Creature — Human {G}, 1/1 (The Dark, #5638):
//
//	"{G}, {T}, Sacrifice this creature: Destroy target artifact."
//
// A mana, tap and sacrifice cost with Shatter's target clause.
// Artifact() reads the current types, so an animated artifact is a
// legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74f361f8-a0d8-4824-9380-61469d6a19cf",
		Name:         "Scavenger Folk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
