package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cone of Flame — Sorcery {3}{R}{R} (Jumpstart 2022, #5948):
//
//	"Cone of Flame deals 1 damage to any target, 2 damage to another target,
//	 and 3 damage to a third target."
//
// Arc Trail's shape with a third clause. Three instances of the word
// "target" are three clauses (CR 601.2c), the second and third Distinct
// from every earlier pick ("another", "a third"). Each slot is checked on
// its own as the spell resolves (CR 608.2b): one that left doesn't stop the
// other damage. All three are dealt as one damage instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62879f3d-ed6d-4004-8196-f3fd7f614695",
		Name:         "Cone of Flame",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetAny(),
			effects.Distinct(effects.TargetAny()),
			effects.Distinct(effects.TargetAny()),
		),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 1), effects.DamageToTarget(1, 2), effects.DamageToTarget(2, 3)),
		OnResolve: effects.DamageEachClauseTarget(1, 2, 3),
	})
}
