package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go Forth — Instant {G} (Foundations Jumpstart, #5464):
//
//	"Choose one —
//	 • Search your library for a basic land card, reveal it, put it
//	   into your hand, then shuffle.
//	 • Target creature gets +2/+2 until end of turn."
//
// Bushwhack's land bullet beside Tifa's Limit Break's Somersault: the
// pump bullet's target is its own group (ModeTarget), so a target that
// left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4f3df30-383b-47ad-8576-0ee69349faf5",
		Name:         "Go Forth",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Search your library for a basic land card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Go Forth — a basic land card",
					}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +2/+2 until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Go Forth"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
		),
	})
}
