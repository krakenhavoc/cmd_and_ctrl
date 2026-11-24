package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri, Storm of Stone — Legendary Planeswalker — Nahiri {2}{R/W}{R/W}, loyalty 6:
//
//	"During your turn, creatures you control have first strike and equip
//	 abilities you activate cost {1} less to activate.
//	 −X: Nahiri deals X damage to target tapped creature."
//
// One printed static, two halves: a first-strike grant that applies only while
// it is the controller's turn, and an equip discount read at activation time
// (the discount holds only for activations made during the controller's turn).
// −X is Chandra, Chill of Compliance's shape: X is announced with the
// activation (CR 107.3a, CR 606.6), the target is re-checked as it resolves.
//
// No simplification.
func init() {
	firstStrike := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && effects.IsActivePlayer(g, source.Controller)
	}, "first strike")
	firstStrike.Label = "During your turn, creatures you control have first strike."
	effects.Register(effects.Spec{
		OracleID:        "6332d6d3-1d77-486c-b040-02fc7225f245",
		Name:            "Nahiri, Storm of Stone",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		StartingLoyalty: 6,
		Static:          []game.StaticAbility{firstStrike},
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsLess(1, "During your turn, equip abilities you activate cost {1} less to activate.",
				effects.AnEquipAbilityCost(), effects.YouActivate(), effects.ActivatedDuringYourTurn()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−X: Nahiri deals X damage to target tapped creature.",
			Cost:    effects.LoyaltyMinusX(),
			Targets: effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
			Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || ctx.X() <= 0 {
					return nil
				}
				return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: t.ID, Amount: ctx.X()}.Apply(ctx)
			})),
		}},
	})
}
