package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Huatli, Radiant Champion — Legendary Planeswalker — Huatli {2}{G}{W}, loyalty 3:
//
//	"+1: Put a loyalty counter on Huatli for each creature you control.
//	 −1: Target creature gets +X/+X until end of turn, where X is the number
//	 of creatures you control.
//	 −8: You get an emblem with "Whenever a creature you control enters, you
//	 may draw a card.""
//
// +1 counts the creatures as it resolves, then puts that many loyalty counters
// on her (the +1's own counter was the cost). −1 reads X as it resolves, with
// the target counted if it is still a creature you control. The emblem's draw
// is a "may" asked per creature that enters.
//
// No simplification.
func init() {
	const emblemText = "Whenever a creature you control enters, you may draw a card."
	effects.Register(effects.Spec{
		OracleID:        "5ab5b82a-83e1-44d7-b8c3-9cd7accc1ca4",
		Name:            "Huatli, Radiant Champion",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Huatli, Radiant Champion emblem",
			Text:  emblemText,
			Triggered: []game.TriggeredAbility{
				effects.OptionalDraw(effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
					"Huatli, Radiant Champion emblem — you may draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
					}), "Huatli, Radiant Champion emblem — draw a card?", 1),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Put a loyalty counter on Huatli for each creature you control.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					n := effects.B04CreaturesControlled(g, item.Controller)
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterLoyalty, N: n}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−1: Target creature gets +X/+X until end of turn, where X is the number of creatures you control.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.B04CreaturesControlled(g, item.Controller)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Huatli, Radiant Champion — +X/+X"}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label: "−8: You get an emblem with \"" + emblemText + "\"",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
