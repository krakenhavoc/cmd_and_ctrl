package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani Vengeant — Legendary Planeswalker — Ajani {2}{R}{W}, starting loyalty
// 3 (Shards of Alara, S82 OLD batch 518, #5588):
//
//	"+1: Target permanent doesn't untap during its controller's next untap
//	 step.
//	 −2: Ajani Vengeant deals 3 damage to any target and you gain 3 life.
//	 −7: Destroy all lands target player controls."
//
// +1 is the one-shot marker on a permanent (DoesntUntapNextUntapStep, CR 502.3),
// whoever controls it. −2 is Essence Drain's two steps (the life gain waits for
// any prompt the damage asks, and with its only target gone nothing happens).
// −7 destroys the lands the target player controls as one sweep, read as the
// ability resolves; an indestructible land survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f953afdf-96f7-4851-994c-1617e8f5e5c9",
		Name:            "Ajani Vengeant",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Target permanent doesn't untap during its controller's next untap step.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("target permanent"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Ajani Vengeant — doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Ajani Vengeant deals 3 damage to any target and you gain 3 life.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetAny(),
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), LifeGain: 3},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ResolveSteps(effects.DamageTheClauseTarget(0, 3), effects.GainLife{Amount: 3})(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−7: Destroy all lands target player controls.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetPlayer("target player"),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsLand() && c.Controller == player
					}}.Apply(ctx)
				},
			},
		},
	})
}
