package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "60754d4b-2add-4928-8d30-8ce9c0f54f91",
		Name:         "Raft Security Officer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}: Tap target creature. This ability costs {1} less to activate if it targets a creature with power 3 or less.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			CostModifiers: []game.CostModifier{
				effects.CostsLessIfItTargets(1, "This ability costs {1} less to activate if it targets a creature with power 3 or less.", effects.And(effects.Creature(), effects.PowerLE(3))),
			},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
