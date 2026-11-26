package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Cathar — Creature — Human Soldier {1}{G}, 2/2 (Shadows over
// Innistrad, S82 OLD):
//
//	"{3}{W}: Target Human gains double strike until end of turn."
//
// Air Marshal's grant over a Human target, the Cathar itself included. A
// target that left or stopped being a Human in response gains nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1eb89059-b691-431d-bcda-ab8aa14c67c8",
		Name:         "Veteran Cathar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Target Human gains double strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")),
			Cost:    effects.ManaCost("{3}{W}"),
			Targets: effects.TargetCreature("target Human", effects.OfCreatureType("Human")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("double strike", "Veteran Cathar — double strike until end of turn"),
		}},
	})
}
