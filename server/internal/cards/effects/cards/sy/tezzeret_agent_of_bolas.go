package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tezzeret, Agent of Bolas — Legendary Planeswalker — Tezzeret {2}{U}{B},
// loyalty 3 (Secret Lair Drop: Planeswalkers, S82 old sets):
//
//	"+1: Look at the top five cards of your library. You may reveal an
//	 artifact card from among them and put it into your hand. Put the rest
//	 on the bottom of your library in any order.
//	 −1: Target artifact becomes an artifact creature with base power and
//	 toughness 5/5.
//	 −4: Target player loses X life and you gain X life, where X is twice
//	 the number of artifacts you control."
//
// −1 has no duration, so it lasts (CR 611.2a), pinned to the object it
// targeted (CR 400.7). X is counted as the −4 resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ab7358e-a0f5-4220-8769-25177237bcdf",
		Name:            "Tezzeret, Agent of Bolas",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Look at the top five cards of your library. You may reveal an artifact card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
						Match:    effects.Artifact(),
						Max:      1,
						Optional: true,
						Label:    "Tezzeret, Agent of Bolas — you may reveal an artifact card and put it into your hand",
						Key:      effects.TakeRevealRestBottomAny,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−1: Target artifact becomes an artifact creature with base power and toughness 5/5.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BecomeArtifactCreature{Target: id, SetPower: 5, SetToughness: 5,
						Label: "Tezzeret, Agent of Bolas — becomes a 5/5 artifact creature"}.Apply(ctx)
				},
			},
			{
				Label:   "−4: Target player loses X life and you gain X life, where X is twice the number of artifacts you control.",
				Cost:    effects.LoyaltyCost(-4),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					who, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					x := 2 * effects.CountYouControl(ctx, effects.Artifact())
					if x <= 0 {
						return nil
					}
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), who, -x); err != nil {
						return err
					}
					return effects.GainLife{Amount: x}.Apply(ctx)
				},
			},
		},
	})
}
