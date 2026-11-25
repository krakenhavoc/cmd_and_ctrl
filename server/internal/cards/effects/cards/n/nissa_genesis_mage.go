package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa, Genesis Mage — {5}{G}{G} Legendary Planeswalker — Nissa (Hour of Devastation), loyalty 5:
//
//	"+2: Untap up to two target creatures and up to two target lands.
//	 −3: Target creature gets +5/+5 until end of turn.
//	 −10: Look at the top ten cards of your library. You may put any number of
//	 creature and/or land cards from among them onto the battlefield. Put the
//	 rest on the bottom of your library in a random order."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac905661-37c4-495a-ac76-96032c23a144",
		Name:            "Nissa, Genesis Mage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Untap up to two target creatures and up to two target lands.",
				Cost:  effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2).Then(
					effects.TargetPermanent("up to two target lands", effects.Land()).WithCount(0, 2)),
				Effect: effects.UntapEachLegalTarget,
			},
			{
				Label:   "−3: Target creature gets +5/+5 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5)),
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 5, Toughness: 5, Label: "Nissa, Genesis Mage — +5/+5"}.Apply(ctx)
				},
			},
			{
				Label: "−10: Look at the top ten cards of your library. You may put any number of creature and/or land cards from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
				Cost:  effects.LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 10),
						Match:    effects.Or(effects.Creature(), effects.Land()),
						Max:      0,
						Optional: true,
						Label:    "Nissa, Genesis Mage — you may put any number of creature and/or land cards onto the battlefield",
						ThenTail: effects.PutRestOnBottomInRandomOrderTail,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
