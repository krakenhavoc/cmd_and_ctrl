package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stun Sniper — Creature — Human Archer {R}{W}, 1/1 (Alara Reborn,
// #5586):
//
//	"{1}, {T}: This creature deals 1 damage to target creature. Tap that
//	 creature."
//
// Centaur Archer's ping (DealDamageToTheTarget), then the tap as a second
// instruction on the same target. Lethal damage is only checked once the
// ability has resolved (CR 704.3), so the creature is tapped either way;
// a target that left in response takes neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5bdf226-7a83-469b-bc49-87af388d9020",
		Name:         "Stun Sniper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This creature deals 1 damage to target creature. Tap that creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.DealDamageToTheTarget(1)(g, item); err != nil {
					return err
				}
				return effects.TapEachLegalTarget(g, item)
			},
		}},
	})
}
