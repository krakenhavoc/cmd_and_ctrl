package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaito Shizuki — {1}{U}{B} Legendary Planeswalker — Kaito, loyalty 3
// (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"At the beginning of your end step, if Kaito entered this turn, he
//	 phases out.
//	 +1: Draw a card. Then discard a card unless you attacked this turn.
//	 −2: Create a 1/1 blue Ninja creature token with "This token can't be
//	 blocked."
//	 −7: You get an emblem with "Whenever a creature you control deals
//	 combat damage to a player, search your library for a blue or black
//	 creature card, put it onto the battlefield, then shuffle.""
//
// The end-step trigger has an intervening "if" (CR 603.4): Kaito must have
// entered this turn as the step begins and as the ability resolves, and he
// phases in during his controller's next untap step (CR 502.1, 702.26). The
// +1 draws, then asks for a discard only when its controller did not attack
// this turn. The Ninja is a catalog token carrying its own can't-be-blocked
// rule (s82_neo_b20_tokens.go). The emblem's trigger fires once per creature
// that connects, under the emblem's owner (CR 114.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba48e067-f4d9-45b2-ade1-71cf8cda418c",
		Name:            "Kaito Shizuki",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Kaito Shizuki emblem",
			Text:  "Whenever a creature you control deals combat damage to a player, search your library for a blue or black creature card, put it onto the battlefield, then shuffle.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
				}, "Kaito Shizuki emblem — search your library for a blue or black creature card, put it onto the battlefield, then shuffle",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						// The colour words are the emblem's text, read through it (CR 612.2).
						blue, black := ctx.TextColor("U"), ctx.TextColor("B")
						return effects.SearchLibrary{
							Player: ctx.Controller(),
							Predicate: func(c game.Card) bool {
								return c.IsCreature() && (c.HasColor(blue) || c.HasColor(black))
							},
							Dest:    game.ZoneBattlefield,
							Limit:   1,
							Shuffle: true,
							Reason:  "Kaito Shizuki emblem — a blue or black creature card, onto the battlefield",
						}.Apply(ctx)
					}))),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(kaitoShizukiEnteredThisTurn,
				effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return g.EnteredThisTurn(source.InstanceID)
				}), "Kaito Shizuki — Kaito phases out",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						return effects.PhaseOut{Targets: []uuid.UUID{ctx.Source()}}.Apply(ctx)
					})))),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card. Then discard a card unless you attacked this turn.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: effects.DrawThenDiscardUnlessAttacked, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				},
			},
			{
				Label:   "−2: Create a 1/1 blue Ninja creature token with \"This token can't be blocked.\"",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.NeoB20NinjaToken(), N: 1}),
			},
			{
				Label:  "−7: You get an emblem with \"Whenever a creature you control deals combat damage to a player, search your library for a blue or black creature card, put it onto the battlefield, then shuffle.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

var kaitoShizukiEnteredThisTurn = effects.CondBoard{
	Name: "Kaito entered this turn",
	Test: func(ctx *effects.Context) bool {
		return effects.OnBattlefield(ctx.Game, ctx.Source()) && ctx.Game.EnteredThisTurn(ctx.Source())
	},
}
