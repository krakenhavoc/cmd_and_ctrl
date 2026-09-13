package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonic Counsel — Sorcery {1}{B} (EDHREC rank 1794):
//
//	"Search your library for a Demon card, reveal it, put it into your
//	 hand, then shuffle.
//	 Delirium — If there are four or more card types among cards in
//	 your graveyard, instead search your library for any card, put it
//	 into your hand, then shuffle."
//
// A two-mana Demonic Tutor with delirium as the price. The card
// types among the caster's graveyard are counted as the spell
// resolves (b16CardTypesInGraveyard, printed type lines); at four or
// more the search is Diabolic Tutor's — any card, unrevealed —
// and below it the search is b06TutorToHand's for a Demon,
// revealed, as printed. Effective subtypes on the library card are
// its printed ones.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "712e3479-722c-40a1-9b61-d5bdde93042b",
		Name:         "Demonic Counsel",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if b16CardTypesInGraveyard(ctx.Game, item.Controller) >= 4 {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(game.Card) bool { return true },
					Dest:      game.ZoneHand,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Demonic Counsel (delirium) — any card, to hand",
				}.Apply(ctx)
			}
			return b06TutorToHand("Demonic Counsel — a Demon card, revealed, to hand",
				func(c game.Card) bool { return c.HasSubtype("Demon") })(item, ctx)
		},
	})
}
