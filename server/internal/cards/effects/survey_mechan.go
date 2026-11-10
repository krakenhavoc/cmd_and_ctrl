package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Survey Mechan — Artifact Creature — Robot {4}, 1/3 (Edge of
// Eternities):
//
//	"Flying
//	 Hexproof
//	 {10}, Sacrifice this creature: It deals 3 damage to any target.
//	 Target player draws three cards and gains 3 life. This ability costs
//	 {X} less to activate, where X is the number of differently named
//	 lands you control."
//
// Astute Terraformer's own-cost clause (CostsLessEach over differently
// named lands, fixed when the cost is determined, CR 601.2f) on a
// {10} generic activation, with the sacrifice paid with it. The ability
// has two targets, "any target" (slot 0) and "target player" (slot 1),
// each judged on its own at resolution (CR 608.2b): losing one leaves
// the other half to do its work. The damage is the Mechan's, as it last
// existed on the battlefield (it was sacrificed as a cost).
//
// No simplification.
func init() {
	const label = "{10}, Sacrifice this creature: It deals 3 damage to any target. Target player draws three cards and gains 3 life. This ability costs {X} less to activate, where X is the number of differently named lands you control."
	Register(Spec{
		OracleID:        "53745903-de8f-42cf-a949-842e5831de30",
		Name:            "Survey Mechan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof"},
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{10}"), SacrificeThis()),
			Targets: TargetAny().Then(TargetPlayer("target player")),
			CostModifiers: []game.CostModifier{
				CostsLessEach(trekB9DifferentlyNamedPermanentsYouControl(Land()),
					"This ability costs {X} less to activate, where X is the number of differently named lands you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if t, ok := ctx.ClauseTarget(0); ok {
					if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
						return err
					}
				}
				if t, ok := ctx.ClauseTarget(1); ok {
					if err := (DrawCards{Player: t.ID, N: 3}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: t.ID, Amount: 3}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
