package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra, Heart of Fire — Legendary Planeswalker — Chandra {3}{R}{R},
// loyalty 5 (Core Set 2021, #5511):
//
//	"+1: Discard your hand, then exile the top three cards of your library.
//	 Until end of turn, you may play cards exiled this way.
//	 +1: Chandra deals 2 damage to any target.
//	 −9: Search your graveyard and library for any number of red instant
//	 and/or sorcery cards, exile them, then shuffle. You may cast them this
//	 turn. Add {R}{R}{R}{R}{R}{R}."
//
// The first +1 discards the hand first, then impulse-exiles three cards with a
// play permission until end of turn (lands included). The −9 searches both
// zones (the searcher names the zones first; the library is shuffled once it
// was searched), takes any number of red instant and sorcery cards to exile,
// and, as the search ends, lets you cast exactly those cards this turn from
// exile (CR 611.2c, paying their costs, at their own timing) and adds six {R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58ad45fc-b625-4484-8d49-ebc42db9040a",
		Name:            "Chandra, Heart of Fire",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Discard your hand, then exile the top three cards of your library. Until end of turn, you may play cards exiled this way.",
				Cost:  effects.LoyaltyCost(1),
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
						return err
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: 3}.Apply(ctx)
					})),
			},
			{
				Label:   "+1: Chandra deals 2 damage to any target.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
				})),
			},
			{
				Label: "−9: Search your graveyard and library for any number of red instant and/or sorcery cards, exile them, then shuffle. You may cast them this turn. Add {R}{R}{R}{R}{R}{R}.",
				Cost:  effects.LoyaltyCost(-9),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.HasColorWord("R") && (c.IsInstant() || c.IsSorcery()) },
						Dest:          game.ZoneExile,
						Unbounded:     true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Chandra, Heart of Fire — any number of red instant and/or sorcery cards",
						ThenTail:      ctx.Tail(chandraHeartOfFireCast),
					}.Apply(ctx)
				},
			},
		},
	})
}

// chandraHeartOfFireCast is the end of the −9: you may cast the exiled cards
// this turn, then add six {R}. An on-disk key.
var chandraHeartOfFireCast = effects.CardTail("then/s82-old-b107-chandra-heart-of-fire-cast", func(ctx *effects.Context, o game.TailOutcome) error {
	var cards []game.Card
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     o.Tail.Player,
		Zone:       game.ZoneExile,
		Duration:   ctx.Game.UntilEndOfTurnDuration(),
		CastOnly:   true,
		Source:     o.Tail.Source,
		SourceName: "Chandra, Heart of Fire",
		Label:      "Chandra, Heart of Fire — cast it this turn",
	}, cards)
	return effects.AddMana{Player: o.Tail.Player, Produced: "{R}{R}{R}{R}{R}{R}"}.Apply(ctx)
})
