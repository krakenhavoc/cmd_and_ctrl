package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// one_mana_tutors.go — the "search your library for a <thing>, reveal
// it, then shuffle and put that card on top" cycle: Enlightened
// Tutor, Worldly Tutor, Mystical Tutor, and Imperial Seal (which
// skips the reveal and charges 2 life instead).
//
// One file rather than four because the difference between them is a
// predicate and a rider, and four files that differ by a one-line
// closure are four places to fix the same bug.
//
// # Why the put-on-top matters
//
// These cost one mana where Demonic Tutor costs three, and the whole
// of the discount is the delay: the card is on TOP of your library,
// not in your hand, so collecting it costs a draw step. Modelling
// them as to-hand searches — which is what the catalog did before
// SearchLibrary.ToTop landed in S22 — prints four cards that are
// strictly better than the ones on the table.
//
// The ordering inside the clause matters too: shuffle FIRST, then
// place. Doing it the other way round shuffles the tutored card back
// into the deck.
//
// # Reveal
//
// Three of the four say "reveal it", and that is a real cost: the
// whole table learns what your next draw is. The engine's reveal
// marks every seat a knower of the card, and because the placement
// happens after the shuffle wiped the zone's knowledge, that
// knowledge is re-granted on the way onto the top. Imperial Seal does
// NOT say reveal — it pays 2 life for the privacy instead.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "c5229c17-b7be-4b05-b683-f2277edc4849",
		Name:     "Enlightened Tutor",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tutorToTop(ctx, "Enlightened Tutor — an artifact or enchantment card", true,
				func(c game.Card) bool { return c.IsArtifact() || c.IsEnchantment() })
		},
	})

	Register(Spec{
		OracleID: "e8863518-0bfa-49c3-8c6e-6c9116a81051",
		Name:     "Worldly Tutor",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tutorToTop(ctx, "Worldly Tutor — a creature card", true,
				func(c game.Card) bool { return c.IsCreature() })
		},
	})

	Register(Spec{
		OracleID: "fb81f95c-70f8-4eb7-8d15-15d0ae23ec03",
		Name:     "Mystical Tutor",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tutorToTop(ctx, "Mystical Tutor — an instant or sorcery card", true,
				func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
		},
	})

	Register(Spec{
		OracleID: "16cd0b90-f70c-4efa-b252-8de8784ef9a3",
		Name:     "Imperial Seal",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// "Search your library for a card, then shuffle and put
			// that card on top. You lose 2 life." No reveal, any
			// card, and the life is a separate sentence that does not
			// wait on the search.
			if err := tutorToTop(ctx, "Imperial Seal — any card", false, nil); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}

// tutorToTop is the shared body: find one card matching `pred` (nil
// means any card), shuffle, put it on top.
func tutorToTop(ctx *Context, reason string, reveal bool, pred func(game.Card) bool) error {
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: pred,
		Dest:      game.ZoneLibrary,
		ToTop:     true,
		Limit:     1,
		Reveal:    reveal,
		Shuffle:   true,
		Reason:    reason,
	}.Apply(ctx)
}
