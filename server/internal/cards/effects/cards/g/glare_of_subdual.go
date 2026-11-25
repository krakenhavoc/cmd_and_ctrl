package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glare of Subdual — Enchantment {2}{G}{W} (Eternal Masters, S82 OLD, #5542):
//
//	"Tap an untapped creature you control: Tap target artifact or creature."
//
// Mothdust Changeling's cost, tapping a creature without the {T} symbol
// (CR 302.6), so one that arrived this turn may pay it; the tap is the
// ability's own effect on its target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcfc112e-0482-4616-8ca9-32338a37fa82",
		Name:         "Glare of Subdual",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: Tap target artifact or creature.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
