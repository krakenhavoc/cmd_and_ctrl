package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vivien, Nature's Avenger — Legendary Planeswalker — Vivien {4}{G}{G}, loyalty 3
// (Core Set 2020, #5518):
//
//	"+1: Put three +1/+1 counters on up to one target creature.
//	 −1: Reveal cards from the top of your library until you reveal a creature
//	 card. Put that card into your hand and the rest on the bottom of your
//	 library in a random order.
//	 −6: Target creature gets +10/+10 and gains trample until end of turn."
//
// The −1 is Spinner of Souls's reveal-until (a library with no creature card
// reveals itself and puts every card on the bottom in a random order).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5513b150-c08f-4f3e-8e41-fd7b43d00b78",
		Name:            "Vivien, Nature's Avenger",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put three +1/+1 counters on up to one target creature.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  effects.OldB155CountersOnTheTarget(3),
			},
			{
				Label:  "−1: Reveal cards from the top of your library until you reveal a creature card. Put that card into your hand and the rest on the bottom of your library in a random order.",
				Cost:   effects.LoyaltyCost(-1),
				Effect: effects.SocB30RevealUntilToHandRestBottom("Vivien, Nature's Avenger", game.Card.IsCreature),
			},
			{
				Label:   "−6: Target creature gets +10/+10 and gains trample until end of turn.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 10, 10, "trample")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 10, Toughness: 10, Label: "Vivien, Nature's Avenger — +10/+10"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Vivien, Nature's Avenger — trample"}.Apply(ctx)
				},
			},
		},
	})
}
