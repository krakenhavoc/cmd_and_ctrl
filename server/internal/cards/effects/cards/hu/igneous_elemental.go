package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Igneous Elemental — Creature — Elemental {4}{R}{R}, 4/3 (Modern Horizons):
//
//	"This spell costs {2} less to cast if there is a land card in your
//	 graveyard.
//	 When this creature enters, you may have it deal 2 damage to target
//	 creature."
//
// The reduction is a self cost modifier (CR 601.2f) that reads the caster's
// graveyard as the spell is cast: any land card there takes {2} off the generic
// part, however many there are. The trigger is Ruination Rioter's shape,
// "you may" asked once the target is chosen; the damage is dealt by the
// Elemental (CR 609.7), and a target that has gone is not hurt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dc8e91a-2c00-4de2-bada-80f323a4ec9d",
		Name:         "Igneous Elemental",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if effects.CountCardsInYourGraveyard(effects.Land())(q) > 0 {
					return 2
				}
				return 0
			}, "This spell costs {2} less to cast if there is a land card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Igneous Elemental — 2 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 2))),
				"Igneous Elemental — have it deal 2 damage to target creature?"),
		},
	})
}
