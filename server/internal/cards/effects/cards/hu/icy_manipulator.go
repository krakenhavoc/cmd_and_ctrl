package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icy Manipulator — Artifact {4} (Dominaria Remastered):
//
//	"{1}, {T}: Tap target artifact, creature, or land."
//
// Avacynian Priest's tap ability over a wider target clause: any
// artifact, creature or land, whoever controls it. A permanent that is
// already tapped is a legal target and stays tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3608f1f7-8dc5-4dd1-ae91-c830e1de9529",
		Name:         "Icy Manipulator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target artifact, creature, or land.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
