package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "c53f2d1c-5a53-4020-aace-a1d9f75e47c7",
		Name:            "Cliffside Rescuer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice this creature: Target permanent you control gains protection from each of your opponents until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerProtect | game.AnswerSacOutlet,
				Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")),
			},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target permanent you control", effects.YouControl()),
			Effect: effects.JmpB7FirstLegalTargetGains("Cliffside Rescuer — protection from each of your opponents",
				game.ProtectionFromEachOfYourOpponents),
		}},
	})
}
