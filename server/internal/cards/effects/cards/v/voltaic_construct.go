package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voltaic Construct — Artifact Creature — Golem Construct {4} (Darksteel, #5605):
//
//	"{2}: Untap target artifact creature."
//
// Aphetto Alchemist's untap with a mana cost and a narrower target: the
// shared untapTheTarget body, the target re-checked as the ability
// resolves (CR 608.2b). It may untap itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5677831-6ca9-4c7b-b02d-f44079c3f663",
		Name:         "Voltaic Construct",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Untap target artifact creature.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
			Effect:  effects.UntapTheTarget,
		}},
	})
}
