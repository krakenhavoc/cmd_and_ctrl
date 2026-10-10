package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra Nalaar — Legendary Planeswalker — Chandra {3}{R}{R},
// starting loyalty 6 (EDHREC rank 14360):
//
//	"+1: Chandra Nalaar deals 1 damage to target player or planeswalker.
//	 −X: Chandra Nalaar deals X damage to target creature.
//	 −8: Chandra Nalaar deals 10 damage to target player or planeswalker
//	     and each creature that player or that planeswalker's controller
//	     controls."
//
// The −X is a loyalty cost of X (LoyaltyMinusX, #1944): X is announced
// with the activation, no more than her loyalty (CR 606.6). The −8's
// damage is one event (CR 120.2): the target and every creature its
// player controls are dealt 10 at once. A planeswalker target names its
// controller as it resolves, the target still being legal then.
func init() {
	Register(Spec{
		OracleID:        "f4da7a43-e9d5-47ea-b2f0-3cf55a156f4a",
		Name:            "Chandra Nalaar",
		Completeness:    CompletenessFull,
		XMatters:        true,
		StartingLoyalty: 6,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Chandra Nalaar deals 1 damage to target player or planeswalker.",
				Cost:    LoyaltyCost(1),
				Targets: targetPlayerOrPlaneswalker(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
				},
			},
			{
				Label:   "−X: Chandra Nalaar deals X damage to target creature.",
				Cost:    LoyaltyMinusX(),
				Targets: TargetCreature("target creature"),
				Purpose: ForTargets(DamageXToTarget(0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok || ctx.X() <= 0 {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}.Apply(ctx)
				},
			},
			{
				Label:   "−8: Chandra Nalaar deals 10 damage to target player or planeswalker and each creature that player or that planeswalker's controller controls.",
				Cost:    LoyaltyCost(-8),
				Targets: targetPlayerOrPlaneswalker(),
				Purpose: ForTargets(DamageToTarget(0, 10)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					player := t.ID
					if t.Kind != game.TargetPlayer {
						c, found := g.LookupCardForEffect(t.ID)
						if !found {
							return nil
						}
						player = c.Controller
					}
					ids := []uuid.UUID{t.ID}
					for _, c := range MatchingBattlefield(ctx, And(Creature(), ControlledBy(player))) {
						if c.InstanceID != t.ID {
							ids = append(ids, c.InstanceID)
						}
					}
					return DealDamageToEachThen(ctx, ids, 10, nil)
				},
			},
		},
	})
}
