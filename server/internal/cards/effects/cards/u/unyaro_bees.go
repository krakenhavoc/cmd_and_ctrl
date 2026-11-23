package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unyaro Bees — Creature — Insect {G}{G}{G}, 0/1 (Time Spiral, #5596):
//
//	"Flying
//	 {G}: This creature gets +1/+1 until end of turn.
//	 {3}{G}, Sacrifice this creature: It deals 2 damage to any target."
//
// Killer Bees' pump (selfPumpRow) and Goblin Legionnaire's sacrifice
// shot: the sacrifice is a cost, so the Bees deal the damage as they
// last existed on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7baff5c8-5919-4cbf-961e-962e740ebd8f",
		Name:            "Unyaro Bees",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.SelfPumpRow("Unyaro Bees", "{G}", 1, 1),
			{
				Label:   "{3}{G}, Sacrifice this creature: It deals 2 damage to any target.",
				Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.B33DamageChosenTargetFromSource(2),
			},
		},
	})
}
