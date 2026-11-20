package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "882619f7-69b4-4ce0-be4e-1e34e608f925",
		Name:            "Chandra, Novice Pyromancer",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Elementals you control get +2/+0 until end of turn.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Match: And(Creature(), Subtype("Elemental"), YouControl()), Power: 2,
						Label: "Chandra, Novice Pyromancer — Elementals you control get +2/+0"}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−1: Add {R}{R}.",
				Cost:  LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddMana{Produced: "{R}{R}"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−2: Chandra deals 2 damage to any target.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 2)),
				Effect:  DealDamageToTheTarget(2),
			},
		},
	})
}
