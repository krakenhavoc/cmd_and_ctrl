package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Finale of Devastation — Sorcery {X}{G}{G} (EDHREC rank ~360):
//
//	"Search your library and/or graveyard for a creature card with
//	 mana value X or less and put it onto the battlefield. If you
//	 search your library this way, shuffle. If X is 10 or more,
//	 creatures you control get +X/+X and gain haste until end of
//	 turn."
//
// A green finisher X-spell: a Craterhoof Behemoth stapled to a
// Chord-of-Calling-shaped tutor, once X clears 10.
//
// # The search covers both zones (#3217)
//
// SearchLibrary{AlsoGraveyard: true}: the picker offers the matching
// creature cards from the library and the graveyard together, marked
// by zone. "And/or" lets the player choose their zones, so the search
// may still find nothing even with a creature in the graveyard (it is a
// search for a stated quality, CR 701.23b). A card taken from the
// graveyard means the library was never searched, so "If you search your
// library this way, shuffle" does not shuffle it; taking a library card
// shuffles. Declared simplification: finding nothing counts as a library
// search (shuffle), the common case; a player who looked only in the
// graveyard and found nothing would not have shuffled, which changes
// nothing in play. Whichever zone it comes from, the pump below still
// runs. See ADR 0015's 2026-10-13 amendment.
//
// # The pump is unconditional on X, not on the search succeeding
//
// "If X is 10 or more, creatures you control get +X/+X and gain haste
// until end of turn" is not gated on the search finding a card — it
// reads X off the announcement (CR 601.2b) and fires regardless, the
// same ruling Craterhoof Behemoth's own X-count establishes for a
// sibling clause. SearchLibrary's `Then` runs once the search has
// finished either way, so the X>=10 check sits there rather than
// inside anything that depends on `found`.
//
// # The pump itself
//
// BoostUntilEOT + GrantKeywordUntilEOT, over "creatures you control"
// snapshotted at THIS moment (CR 611.2c) — Craterhoof's own two-layer
// split (P/T in layer 7c, the keyword grant in layer 6) applies here
// unchanged, and is why one registry entry still calls two primitives
// rather than one.
func init() {
	Register(Spec{
		OracleID:     "69872a9a-fe54-4e58-940c-89395af71acd",
		Name:         "Finale of Devastation",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			tail := ctx.ItemTail(finaleOfDevastationPump)
			tail.N = x
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				// #3217: "library and/or graveyard".
				AlsoGraveyard: true,
				Reason:        "Finale of Devastation — a creature card with mana value X or less",
				// #4253 slice 4: X and the item ride the tail.
				ThenTail: tail,
			}.Apply(ctx)
		},
	})
}

// finaleOfDevastationPump is "if X is 10 or more, creatures you control get
// +X/+X and gain haste until end of turn" (X is Tail.N). An on-disk key.
var finaleOfDevastationPump = CardTail("finale-of-devastation/pump", func(ctx *Context, o game.TailOutcome) error {
	x := o.Tail.N
	if x < 10 {
		return nil
	}
	yours := And(Creature(), YouControl())
	if err := (BoostUntilEOT{
		Match:     yours,
		Power:     x,
		Toughness: x,
		Label:     "Finale of Devastation — +X/+X",
	}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Match:    yours,
		Keywords: []string{"haste"},
		Label:    "Finale of Devastation — haste",
	}.Apply(ctx)
})
