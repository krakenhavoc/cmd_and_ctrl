package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Samut, the Tested — {2}{R}{G} Legendary Planeswalker — Samut (Hour of Devastation), loyalty 4:
//
//	"+1: Up to one target creature gains double strike until end of turn.
//	 −2: Samut deals 2 damage divided as you choose among one or two targets.
//	 −7: Search your library for up to two creature and/or planeswalker cards,
//	 put them onto the battlefield, then shuffle."
//
// The −2 divides as it is announced (CR 601.2d), one or two targets of any kind.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fffdc2ac-bde4-4e4c-a5bd-0e6c6e49ad91",
		Name:            "Samut, the Tested",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature gains double strike until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike")),
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Samut, the Tested — double strike"}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Samut deals 2 damage divided as you choose among one or two targets.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetAny().WithCount(1, 2).Dividing(effects.Divide(2)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DealDividedDamage(effects.NewContext(g, item))
				},
			},
			{
				Label: "−7: Search your library for up to two creature and/or planeswalker cards, put them onto the battlefield, then shuffle.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsCreature() || c.IsPlaneswalker() },
						Dest:      game.ZoneBattlefield,
						Limit:     2,
						UpTo:      true,
						Shuffle:   true,
						Reason:    "Samut, the Tested — up to two creature and/or planeswalker cards onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
