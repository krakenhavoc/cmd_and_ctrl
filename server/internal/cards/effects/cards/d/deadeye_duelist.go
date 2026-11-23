package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadeye Duelist — Creature — Human Assassin {1}{R}, 1/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Reach
//	 {1}, {T}: This creature deals 1 damage to target opponent."
//
// Fervent Paincaster's pinger with a {1} on the cost and the target narrowed
// to an opponent. The damage is dealt by the creature, and the opponent is
// checked again as the ability resolves (CR 608.2b). Targeting an opponent
// is a crime (CR 700.13), which the engine reads from the activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29ef90ac-2396-4488-970d-d440297ba5ee",
		Name:            "Deadeye Duelist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This creature deals 1 damage to target opponent.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
