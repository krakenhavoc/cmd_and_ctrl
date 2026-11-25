package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nylea's Intervention — Sorcery {X}{G}{G} (Theros Beyond Death, S82 OLD,
// #5514):
//
//	"Choose one —
//	 • Search your library for up to X land cards, reveal them, put them into
//	   your hand, then shuffle.
//	 • Nylea's Intervention deals twice X damage to each creature with
//	   flying."
//
// A two-bullet modal sorcery with X from the cast. The first is a revealed
// search of up to X land cards into the hand (X = 0 finds none and still
// shuffles). The second is a sweep: twice X damage to every creature that has
// flying as the spell resolves, dealt simultaneously (CR 120.3) by the spell,
// whoever controls them, so it declares the Sweep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acf388b2-c4e3-4f1b-a16c-88f991d5c17b",
		Name:         "Nylea's Intervention",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Search your library for up to X land cards, reveal them, put them into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsLand() },
						Dest:      game.ZoneHand,
						Limit:     ctx.X(),
						UpTo:      true,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Nylea's Intervention — up to X land cards, revealed, to hand",
					}.Apply(ctx)
				}), game.Purpose{Tutors: 1}),
			effects.ModeWithPurpose(effects.ModeDoing("Nylea's Intervention deals twice X damage to each creature with flying.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), 2*ctx.X())
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}}),
		),
	})
}
