package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiler of Instinct — Creature — Phyrexian Kavu {2}{R}{R}, 4/4 (Dominaria United,
// #5659):
//
//	"First strike
//	 As an additional cost to cast red permanent spells, you may pay 2
//	 life. Those spells cost {R} less to cast if you paid life this way.
//	 This effect reduces only the amount of red mana you pay.
//	 Whenever you cast a red permanent spell, this creature deals 1 damage
//	 to any target."
//
// The offer and its discount are the cycle's linked pair
// (pay_life_for_discount.go): each red permanent spell its controller
// casts may be announced with "pay 2 life", and one that is costs {R}
// less, never less generic mana. The Defiler is
// the source of the damage, from last-known information once it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "590137df-af61-4e93-b6f1-3a2e438b8037",
		Name:                 "Defiler of Instinct",
		Completeness:         effects.CompletenessFull,
		PrintedKeywords:      []string{"first strike"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.MayPayLifeToCast("Defiler of Instinct", "R", 2)},
		CostModifiers:        []game.CostModifier{effects.CostsColorLessIfLifePaid("R")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WheneverYouCastAPermanentSpellOfColor("R",
				"Defiler of Instinct — this creature deals 1 damage to any target",
				effects.SourceDealsDamageToEachLegalTarget(1)), effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
