package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ral, Crackling Wit — Legendary Planeswalker — Ral {2}{U}{R}, loyalty 4 (Bloomburrow):
//
//	"Whenever you cast a noncreature spell, put a loyalty counter on Ral.
//	 +1: Create a 1/1 blue and red Otter creature token with prowess.
//	 −3: Draw three cards, then discard two cards.
//	 −10: Draw three cards. You get an emblem with "Instant and sorcery spells
//	 you cast have storm." (Whenever you cast an instant or sorcery spell, copy
//	 it for each spell cast before it this turn.)"
//
// The static trigger is a one-sentence "put a counter on this" (CounterOnThis,
// which declares a footprint). The +1 is the table's Otter row (the Stormchaser's
// Talent token). The −3 draws before it asks for the discards (the discard
// prompt waits for any draw prompt, CR 608.2c). The −10 draws, then creates the
// emblem: its storm is the Prismari grant (socB32GrantsStorm), a trigger on the
// emblem over every instant or sorcery spell its owner casts, copying it once
// for each spell cast before it this turn, every player's included (CR 702.40a).
// Nothing removes an emblem (CR 114.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c970f43-17da-4187-a084-efd53b579067",
		Name:         "Ral, Crackling Wit",
		Completeness: effects.CompletenessFull,
		Emblem: &effects.EmblemSpec{
			Label: "Ral, Crackling Wit emblem",
			Text:  "Instant and sorcery spells you cast have storm.",
			Triggered: []game.TriggeredAbility{
				effects.SocB32GrantsStorm("Ral, Crackling Wit emblem", func(spell game.Card, _ *game.Card, _ *game.Game) bool {
					return spell.IsInstant() || spell.IsSorcery()
				}),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Ral, Crackling Wit — put a loyalty counter on Ral",
				effects.Do(effects.CounterOnThis{Kind: game.CounterLoyalty, N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Create a 1/1 blue and red Otter creature token with prowess.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue and red Otter with prowess"), N: 1}),
			},
			{
				Label:   "−3: Draw three cards, then discard two cards.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Draws: 3, Discards: 2},
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 3}.Apply(ctx) }),
					effects.Instruction(func(ctx *effects.Context) error {
						ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 2})
						return nil
					}),
				),
			},
			{
				Label:   "−10: Draw three cards. You get an emblem with \"Instant and sorcery spells you cast have storm.\"",
				Cost:    effects.LoyaltyCost(-10),
				Purpose: game.Purpose{Draws: 3},
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 3}.Apply(ctx) }),
					effects.Instruction(func(ctx *effects.Context) error { return effects.CreateEmblem{}.Apply(ctx) }),
				),
			},
		},
	})
}
