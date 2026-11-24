package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Flame's Catalyst — Legendary Planeswalker — Chandra {4}{R}{R},
// loyalty 5 (Secret Lair Drop: Planeswalkers, S82 old sets):
//
//	"+1: Chandra deals 3 damage to each opponent.
//	 −2: You may cast target red instant or sorcery card from your graveyard.
//	 If that spell would be put into your graveyard, exile it instead.
//	 −8: Discard your hand, then draw seven cards. Until end of turn, you
//	 may cast spells from your hand without paying their mana costs."
//
// −2 is the cast offered during resolution at the card's mana cost (CR
// 608.2g, OfferGraveyardCast, Emet-Selch's shape): the spell is exiled if it
// would go to the graveyard however it gets there, and an uncast card stays
// where it is.
//
// −8's free casting is a standing, never-spent promise held on the player
// until end of turn (NextSpellPromise EveryMatchingSpell + WithoutPayingManaCost
// + FromHandOnly): spells cast from anywhere but the hand still cost mana.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7721e800-fba6-4ae7-855e-631b2ecc8d6b",
		Name:            "Chandra, Flame's Catalyst",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Chandra deals 3 damage to each opponent.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.DamageEachOpponent{N: 3}),
			},
			{
				Label: "−2: You may cast target red instant or sorcery card from your graveyard. If that spell would be put into your graveyard, exile it instead.",
				Cost:  effects.LoyaltyCost(-2),
				Targets: effects.TargetCardInGraveyard("target red instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.OfColor("R"), effects.Or(effects.Instant(), effects.Sorcery())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.OfferGraveyardCast{Target: id, SourceName: "Chandra, Flame's Catalyst"}.Apply(ctx)
				},
			},
			{
				Label:   "−8: Discard your hand, then draw seven cards. Until end of turn, you may cast spells from your hand without paying their mana costs.",
				Cost:    effects.LoyaltyCost(-8),
				Purpose: game.Purpose{Draws: 7},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if _, err := effects.DiscardWholeHand(g, item.Controller); err != nil {
						return err
					}
					if err := (effects.DrawCards{N: 7}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantNextSpellPromise{From: "Chandra, Flame's Catalyst", Promise: game.NextSpellPromise{
						EveryMatchingSpell:    true,
						WithoutPayingManaCost: true,
						FromHandOnly:          true,
						Text:                  "Until end of turn, you may cast spells from your hand without paying their mana costs.",
					}}.Apply(ctx)
				},
			},
		},
	})
}
