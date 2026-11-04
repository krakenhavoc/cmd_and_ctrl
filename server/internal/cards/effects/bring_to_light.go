package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bring to Light — Sorcery {3}{G}{U} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Converge — Search your library for a creature, instant, or sorcery
//	 card with mana value less than or equal to the number of colors of
//	 mana spent to cast this spell, exile that card, then shuffle. You
//	 may cast that card without paying its mana cost."
//
// Converge is an ability word, so the counting lives here. Like Painful
// Truths it declares ReadsManaSpent (a waived payment asks what was paid
// outside the pool) and WantsDistinctColors (the generic half of the cost
// is paid with colours not yet spent), and it reads
// ctx.ColorsSpentCount() as the search starts. The search takes one card
// that is a creature, an instant or a sorcery with a mana value no
// higher than that count, to exile, then shuffles; a card with {X} in its
// cost is 0 there (CR 202.3e), and a split card is the sum of its halves
// (CR 202.3d).
//
// "You may cast that card" is the free-cast offer a resolution makes
// (Emergent Ultimatum's, CR 608.2g, #4250): a {0} cast during the
// resolution, timing ignored, so a sorcery can be cast at once. Declined,
// the card stays in exile. A search that finds nothing offers nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "67183256-23cc-4af5-991e-d7148e942c63",
		Name:                "Bring to Light",
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Completeness:        CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			colors := ctx.ColorsSpentCount()
			return SearchLibrary{
				Player: ctx.Controller(),
				Limit:  1,
				Dest:   game.ZoneExile,
				Predicate: func(c game.Card) bool {
					return (c.HasCardType("creature") || c.HasCardType("instant") || c.HasCardType("sorcery")) &&
						c.ManaValue() <= colors
				},
				Shuffle:  true,
				Reason:   "Bring to Light — search for a creature, instant or sorcery card with mana value " + strconv.Itoa(colors) + " or less",
				ThenTail: ctx.ItemTail(bringToLightCast),
			}.Apply(ctx)
		},
	})
}

// bringToLightCast is "you may cast that card without paying its mana
// cost" over the card the search exiled. An on-disk key.
var bringToLightCast = CardTail("bring-to-light/cast", func(ctx *Context, o game.TailOutcome) error {
	grantFreeCasts(ctx.Game, ctx.Controller(), ctx.Source(), "Bring to Light",
		game.TimingFlash, game.LapseStaysInExile, o.IDs)
	return nil
})
