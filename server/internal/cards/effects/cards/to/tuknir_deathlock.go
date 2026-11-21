package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tuknir Deathlock — Legendary Creature — Human Wizard {R}{R}{G}{G}, 2/2
// (Legends, #5639):
//
//	"Flying
//	 {R}{G}, {T}: Target creature gets +2/+2 until end of turn."
//
// Alaborn Veteran's pump (boostTargetUntilEOT) with a mana cost beside
// the tap and no timing window: any creature, at instant speed. The tap
// means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b6fa658-a344-4202-866b-e19997e32ff2",
		Name:            "Tuknir Deathlock",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label: "{R}{G}, {T}: Target creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			},
			Cost:    effects.Plus(effects.ManaCost("{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Tuknir Deathlock — +2/+2", 2, 2),
		}},
	})
}
