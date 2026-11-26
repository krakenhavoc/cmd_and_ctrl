package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naya Battlemage — Creature — Human Shaman {2}{G}, 2/2 (Shards of Alara,
// S82 OLD batch 511, #5588):
//
//	"{R}, {T}: Target creature gets +2/+0 until end of turn.
//	 {W}, {T}: Tap target creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db4d2580-50f4-4986-8091-769b4d2f6d72",
		Name:         "Naya Battlemage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}: Target creature gets +2/+0 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Naya Battlemage — +2/+0 until end of turn", 2, 0),
			},
			{
				Label:   "{W}, {T}: Tap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.ThousandMoonsCrackshotTap,
			},
		},
	})
}
