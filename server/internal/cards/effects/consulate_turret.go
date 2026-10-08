package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consulate Turret — Artifact {3}:
//
//	"{T}: You get {E} (an energy counter).
//	 {T}, Pay {E}{E}{E}: This artifact deals 2 damage to target player or
//	 planeswalker."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14d71432-c2e6-4c81-92ef-cf98e1fcf5f8",
		Name:         "Consulate Turret",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: You get {E}.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Energy: 1},
				Effect:  ebYouGetEnergy(1),
			},
			{
				Label:   "{T}, Pay {E}{E}{E}: This artifact deals 2 damage to target player or planeswalker.",
				Cost:    Plus(TapCost(), PayEnergy(3)),
				Targets: targetPlayerOrPlaneswalker(),
				Purpose: ForTargets(DamageToTarget(0, 2)),
				Effect:  sourceDealsDamageToEachLegalTarget(2),
			},
		},
	})
}
