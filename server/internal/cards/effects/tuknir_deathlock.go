package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0b6fa658-a344-4202-866b-e19997e32ff2",
		Name:            "Tuknir Deathlock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{R}{G}, {T}: Target creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2)),
			},
			Cost:    Plus(ManaCost("{R}{G}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Tuknir Deathlock — +2/+2", 2, 2),
		}},
	})
}
