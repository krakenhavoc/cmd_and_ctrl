package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fires of Yavimaya — Enchantment {1}{R}{G} (Invasion, #5616):
//
//	"Creatures you control have haste.
//	 Sacrifice this enchantment: Target creature gets +2/+2 until end
//	 of turn."
//
// Fervor's haste anthem (a layer-6 TribalKeywordGrant with no tribe,
// yours only), and a sacrifice ability with the shared
// boostTargetUntilEOT body. The sacrifice is the whole cost, so the
// pump is at instant speed and any creature may be the target. Once the
// enchantment is gone its haste grant goes with it, so a creature that
// entered this turn and has not yet attacked loses haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e23d6f3b-0e18-423b-943b-15db7837255b",
		Name:         "Fires of Yavimaya",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "haste"),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice this enchantment: Target creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump,
				Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2)),
			},
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Fires of Yavimaya — +2/+2", 2, 2),
		}},
	})
}
