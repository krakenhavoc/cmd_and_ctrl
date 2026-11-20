package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorched Rusalka — Creature — Spirit {R}, 1/1 (Guildpact, #5599):
//
//	"{R}, Sacrifice a creature: This creature deals 1 damage to target
//	 player or planeswalker."
//
// Fervent Paincaster's ping behind High Market's sacrifice: any creature
// its controller controls pays the cost. If the Rusalka sacrificed
// itself it still deals the damage, as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd05fd8e-f25e-4b9e-b162-afea1b248282",
		Name:         "Scorched Rusalka",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice a creature: This creature deals 1 damage to target player or planeswalker.",
			Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 1)), Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{R}"), SacrificeACreature()),
			Targets: targetPlayerOrPlaneswalker(),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
