package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Novice Pyromancer — Legendary Planeswalker — Chandra {3}{R}, starting loyalty 5 (Core Set 2020, #5518):
//
//	"+1: Elementals you control get +2/+0 until end of turn.
//	 −1: Add {R}{R}.
//	 −2: Chandra deals 2 damage to any target."
//
// The +1's set of Elementals is fixed as it resolves (CR 611.2c). The −1
// is a loyalty ability that adds mana, so it uses the stack (CR 606.3,
// Chandra, Torch of Defiance's +1). The −2 is the shared target damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "882619f7-69b4-4ce0-be4e-1e34e608f925",
		Name:            "Chandra, Novice Pyromancer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Elementals you control get +2/+0 until end of turn.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.Subtype("Elemental"), effects.YouControl()), Power: 2,
						Label: "Chandra, Novice Pyromancer — Elementals you control get +2/+0"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label: "−1: Add {R}{R}.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AddMana{Produced: "{R}{R}"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−2: Chandra deals 2 damage to any target.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.DealDamageToTheTarget(2),
			},
		},
	})
}
