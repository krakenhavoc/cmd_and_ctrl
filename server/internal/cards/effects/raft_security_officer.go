package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raft Security Officer — Creature — Human Soldier {1}{W}, 1/3 (Marvel
// Super Heroes, #4455):
//
//	"{2}, {T}: Tap target creature. This ability costs {1} less to
//	 activate if it targets a creature with power 3 or less."
//
// Arashin Sunshield's tap ability with the ability's own cost clause
// (ActivatedAbility.CostModifiers, #1296): CostsLessIfItTargets prices it
// from the announced target as it stands when the cost is determined
// (CR 602.2b, 601.2f), so a target of power 3 or less makes it {1},{T} and
// anything bigger the printed {2},{T}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60754d4b-2add-4928-8d30-8ce9c0f54f91",
		Name:         "Raft Security Officer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Tap target creature. This ability costs {1} less to activate if it targets a creature with power 3 or less.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessIfItTargets(1, "This ability costs {1} less to activate if it targets a creature with power 3 or less.", And(Creature(), PowerLE(3))),
			},
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
