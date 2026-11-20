package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f5bdf226-7a83-469b-bc49-87af388d9020",
		Name:         "Stun Sniper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: This creature deals 1 damage to target creature. Tap that creature.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := DealDamageToTheTarget(1)(g, item); err != nil {
					return err
				}
				return tapEachLegalTarget(g, item)
			},
		}},
	})
}
