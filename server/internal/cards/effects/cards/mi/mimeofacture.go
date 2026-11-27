package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mimeofacture — Sorcery {3}{U} (Guildpact, S82 OLD batch 573, #5599):
//
//	"Replicate {3}{U} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.)
//	 Choose target permanent an opponent controls. Search that player's
//	 library for a card with the same name and put it onto the battlefield
//	 under your control. Then that player shuffles."
//
// Replicate is the engine's keyword (replicate.go, CR 702.56a): an optional
// {3}{U} paid any number of times, the engine's cap of 20 payments on one
// announcement being no table's limit. The search is a stated-quality search
// (CR 701.23b) of the target's controller's library by the caster, the cards
// entering under the caster's control; the name is the permanent's as it
// last existed (CR 608.2h), through both halves of a double-faced card. The
// library's owner shuffles whether or not a card was found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "79b978b8-44c8-4ca8-bdd5-10490145398d",
		Name:          "Mimeofacture",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPermanent("target permanent an opponent controls", effects.OpponentControls()),
		OptionalCosts: []game.AdditionalCost{effects.ReplicateMana("{3}{U}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: c.Controller,
				Predicate:    effects.SharesANameWith(effects.ExtractionNames(c)),
				Dest:         game.ZoneBattlefield,
				Limit:        1,
				Shuffle:      true,
				Reason:       "Mimeofacture — search the library for a card with the same name",
			}.Apply(ctx)
		},
	})
}
