package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cliffside Rescuer — Creature — Kor Soldier {1}{W}, 2/2 (Commander 2019,
// #5517):
//
//	"Vigilance
//	 {T}, Sacrifice this creature: Target permanent you control gains
//	 protection from each of your opponents until end of turn."
//
// The tap and the sacrifice are costs (CR 602.2b), so the Rescuer is gone
// before anyone can respond and may not protect itself. The grant is the
// printed "protection from each of your opponents" quality (Absolute
// Virtue's, CR 702.16k) as an until-end-of-turn layer 6 keyword on the
// target, which judges each source by who controls it. A target that left
// in response gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c53f2d1c-5a53-4020-aace-a1d9f75e47c7",
		Name:            "Cliffside Rescuer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice this creature: Target permanent you control gains protection from each of your opponents until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerProtect | game.AnswerSacOutlet,
				Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "protection")),
			},
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target permanent you control", YouControl()),
			Effect: jmpB7FirstLegalTargetGains("Cliffside Rescuer — protection from each of your opponents",
				game.ProtectionFromEachOfYourOpponents),
		}},
	})
}
