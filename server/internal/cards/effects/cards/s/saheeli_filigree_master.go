package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saheeli, Filigree Master — Legendary Planeswalker — Saheeli {2}{U}{R},
// loyalty 3 (The Brothers' War, #5482):
//
//	"+1: Scry 1. You may tap an untapped artifact you control. If you do,
//	     draw a card.
//	 −2: Create two 1/1 colorless Thopter artifact creature tokens with
//	     flying. They gain haste until end of turn.
//	 −4: You get an emblem with "Artifact creatures you control get +1/+1"
//	     and "Artifact spells you cast cost {1} less to cast.""
//
// The +1 is Scry 1 with the rest of the ability as its tail, so the artifact
// is asked about only once the scry is done; the tap and the card are one
// choice (a chosen artifact that is no longer untapped draws nothing). The
// −2 grants haste to the two Thopters that landed, until end of turn.
//
// The −4 emblem has two abilities. The pump is a static over artifact
// creatures its owner controls. The discount is the emblem's own cost
// modifier (EmblemSpec.CostModifiers, priced like any other, CR 601.2f).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68cf7415-b434-4284-89b4-c7d44a003eb8",
		Name:            "Saheeli, Filigree Master",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Saheeli, Filigree Master emblem",
			Text:  "Artifact creatures you control get +1/+1. Artifact spells you cast cost {1} less to cast.",
			Static: []game.StaticAbility{{
				Label:    "Artifact creatures you control get +1/+1.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			}},
			CostModifiers: []game.CostModifier{
				effects.CostsLess(1, "Artifact spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ArtifactSpell()),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Scry 1. You may tap an untapped artifact you control. If you do, draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Scry: 1, Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{N: 1, ThenTail: ctx.ItemTail(effects.S82BroB20SaheeliAfterScry)}.Apply(ctx)
				},
			},
			{
				Label: "−2: Create two 1/1 colorless Thopter artifact creature tokens with flying. They gain haste until end of turn.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("1/1 colorless Thopter artifact with flying")), Count: 2}},
					}, ctx.ItemTail(effects.S82BroB20ThopterHaste))
				},
			},
			{
				Label:  "−4: You get an emblem with \"Artifact creatures you control get +1/+1\" and \"Artifact spells you cast cost {1} less to cast.\"",
				Cost:   effects.LoyaltyCost(-4),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
