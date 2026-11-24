package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska, Golgari Queen — Legendary Planeswalker — Vraska {2}{B}{G}, loyalty 4:
//
//	"+2: You may sacrifice another permanent. If you do, you gain 1 life and
//	 draw a card.
//	 −3: Destroy target nonland permanent with mana value 3 or less.
//	 −9: You get an emblem with "Whenever a creature you control deals combat
//	 damage to a player, that player loses the game.""
//
// +2 asks the sacrifice (never Vraska herself); the life and the card are the
// "if you do" and happen only for a permanent that really went. The emblem's
// trigger fires once per creature that dealt combat damage to a player, and
// the player it names is the one damaged.
//
// No simplification.
func init() {
	const emblemText = "Whenever a creature you control deals combat damage to a player, that player loses the game."
	effects.Register(effects.Spec{
		OracleID:        "131096a3-2d53-4ec2-96d3-fcb58e698a13",
		Name:            "Vraska, Golgari Queen",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Vraska, Golgari Queen emblem",
			Text:  emblemText,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventDealDamage,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
					},
					"Vraska, Golgari Queen emblem — that player loses the game",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
					}),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: You may sacrifice another permanent. If you do, you gain 1 life and draw a card.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Vraska, Golgari Queen — sacrifice another permanent to gain 1 life and draw a card?",
						effects.Permanent(), true, vraskaGolgariQueenSacrificed)
				},
			},
			{
				Label:   "−3: Destroy target nonland permanent with mana value 3 or less.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target nonland permanent with mana value 3 or less", effects.Nonland(), effects.ManaValueLE(3)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			},
			{
				Label:  "−9: You get an emblem with \"" + emblemText + "\"",
				Cost:   effects.LoyaltyCost(-9),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// vraskaGolgariQueenSacrificed is the "if you do". An on-disk key: never
// renamed, never reused.
var vraskaGolgariQueenSacrificed = effects.RegisterSacrificedOne("prompt-run/vraska-golgari-queen-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1})(ctx.Game, ctx.Item)
})
