package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slick Imitator — Creature — Ooze {1}{U}, 1/3:
//
//	"Start your engines!
//	 Max speed — {1}, Sacrifice this creature: Copy target spell you
//	 control. You may choose new targets for the copy. (A copy of a
//	 permanent spell becomes a token.)"
//
// ADR 0138 (#2711). The target is any spell its controller controls; the
// sacrifice is paid at announce, so the copy resolves without the
// Imitator. Lithoform Engine's copy body, with the CR 707.10c re-target
// offer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5482ddb1-7f7d-492a-aa3d-67c037062cc3",
		Name:            "Slick Imitator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:   "Max speed — {1}, Sacrifice this creature: Copy target spell you control. You may choose new targets for the copy.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
				Targets: effects.TargetSpell("target spell you control", effects.YouControl()),
				Effect:  effects.CopyTargetedSpell,
			}),
		},
	})
}
