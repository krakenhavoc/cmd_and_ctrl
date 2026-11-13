package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a4f3df30-383b-47ad-8576-0ee69349faf5",
		Name:         "Go Forth",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Search your library for a basic land card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Go Forth — a basic land card",
					}.Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Target creature gets +2/+2 until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Go Forth"}.Apply(ctx)
				}), ForTargets(BenefitUntilEOT(0, 2, 2))),
		),
	})
}
