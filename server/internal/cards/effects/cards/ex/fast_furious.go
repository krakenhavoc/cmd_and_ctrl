package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fast // Furious — split card (Modern Horizons 2, #5501):
//
//	Fast — Instant {2}{R}: "Discard a card, then draw two cards."
//	Furious — Sorcery {3}{R}{R}: "Furious deals 3 damage to each creature
//	          without flying."
//
// Each half is an ADR 0034 face (ADR 0103): Fast under the bare oracle ID,
// Furious under "#1". Only one half is cast (CR 709.3; the card has no
// fuse). Fast is a rummage with a two-card draw, the draw keyed behind the
// discard; with an empty hand nothing is discarded and the two cards are
// still drawn. Furious is Rolling Temblor's wipe at 3, sparing fliers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62411ced-843e-4b63-bdf6-dafb2ac27047",
		Name:         "Fast",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: ctx.Controller(),
				Source: item.SourceCardID,
				N:      1,
			}, effects.OldB029DrawTwoAfterDiscard, nil)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "62411ced-843e-4b63-bdf6-dafb2ac27047#1",
		Name:         "Furious",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 3)
		},
	})
}
