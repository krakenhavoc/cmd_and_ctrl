package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fervent Paincaster — Creature — Human Wizard {2}{R}, 3/1:
//
//	"{T}: This creature deals 1 damage to target player or planeswalker.
//	 {T}, Exert this creature: It deals 1 damage to target creature.
//	 (An exerted creature won't untap during your next untap step.)"
//
// Two pingers; the creature one pays an exert as well as the {T}
// (ADR 0130 §4, CR 701.43a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d569df1-23cf-4e01-8ef9-a1a8b815f11e",
		Name:         "Fervent Paincaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: This creature deals 1 damage to target player or planeswalker.",
				Cost:    TapCost(),
				Targets: targetPlayerOrPlaneswalker(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			},
			{
				Label:   "{T}, Exert this creature: It deals 1 damage to target creature.",
				Cost:    Plus(TapCost(), ExertThis()),
				Targets: TargetCreature("target creature"),
				Purpose: game.Purpose{DamageToCreature: 1},
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			},
		},
	})
}
