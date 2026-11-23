package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret, Artifice Master — Legendary Planeswalker {3}{U}{U}, loyalty 5
// (Jumpstart 2022, #5948):
//
//	"+1: Create a 1/1 colorless Thopter artifact creature token with
//	 flying.
//	 0: Draw a card. If you control three or more artifacts, draw two
//	 cards instead.
//	 −9: You get an emblem with "At the beginning of your end step,
//	 search your library for a permanent card, put it onto the
//	 battlefield, then shuffle.""
//
// The 0 counts the artifacts as it resolves and draws one or two. The
// emblem triggers from the command zone at its owner's end step, and the
// search is for any permanent card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e22824cf-07a1-4c83-b6c4-9d8fcff3892f",
		Name:            "Tezzeret, Artifice Master",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Tezzeret, Artifice Master emblem",
			Text:  "At the beginning of your end step, search your library for a permanent card, put it onto the battlefield, then shuffle.",
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Tezzeret, Artifice Master emblem — search your library for a permanent card",
					func(g *game.Game, item *game.StackItem) error {
						return effects.SearchLibrary{
							Player:    item.Controller,
							Predicate: func(c game.Card) bool { return c.IsPermanent() },
							Dest:      game.ZoneBattlefield,
							Limit:     1,
							Shuffle:   true,
							Reason:    "Tezzeret, Artifice Master emblem — search your library for a permanent card",
						}.Apply(effects.NewContext(g, item))
					}),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a 1/1 colorless Thopter artifact creature token with flying.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
			},
			{
				Label:   "0: Draw a card. If you control three or more artifacts, draw two cards instead.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Draws: 1},
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := 1
					if len(effects.MatchingBattlefield(ctx, effects.And(effects.Artifact(), effects.YouControl()))) >= 3 {
						n = 2
					}
					return effects.DrawCards{N: n}.Apply(ctx)
				})),
			},
			{
				Label: "−9: You get an emblem with \"At the beginning of your end step, search your library for a permanent card, put it onto the battlefield, then shuffle.\"",
				Cost:  effects.LoyaltyCost(-9),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
