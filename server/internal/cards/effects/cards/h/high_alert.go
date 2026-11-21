package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High Alert — Enchantment {1}{W}{U} (The List):
//
//	"Each creature you control assigns combat damage equal to its toughness
//	 rather than its power.
//	 Creatures you control can attack as though they didn't have defender.
//	 {2}{W}{U}: Untap target creature."
//
// The first ability is Assault Formation's layer-6 flag (CR 510.1c), the
// second Guardians of Oboro's defender waiver aimed at every creature its
// controller has, and the activation untaps any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f657279-cc0f-4431-b517-47bd1c7109b6",
		Name:         "High Alert",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer: game.Layer6Ability,
				Label: "Each creature you control assigns combat damage equal to its toughness rather than its power.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.DamageByToughnessAlways = true
				},
			},
			highAlertDefenderWaiver(),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{U}: Untap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{W}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.UntapTheTarget,
		}},
	})
}

func highAlertDefenderWaiver() game.StaticAbility {
	ab := effects.AvB18CanAttackAsThoughNoDefenderWhile(
		"Creatures you control can attack as though they didn't have defender.",
		func(*game.Game, *game.Card) bool { return true })
	ab.AppliesTo = func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller
	}
	return ab
}
