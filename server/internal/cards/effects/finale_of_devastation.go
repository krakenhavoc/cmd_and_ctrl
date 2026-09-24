package effects

import (
	"github.com/google/uuid"

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
// # Declared simplification: library only (Tower Winder's shape)
//
// SearchLibrary is a library-zone primitive; the engine has no
// combined library-AND-graveyard search yet (the same gap Tower
// Winder's own header documents). A creature card already in the
// graveyard can't be found here. The loss is real but bounded: the
// library is where the target usually is, this never finds something
// the printed card couldn't, and the shuffle still fires exactly as
// printed for the half that IS implemented.
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
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The search looks in your library only. A creature card already in your graveyard can't be found.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			item := ctx.Item
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Finale of Devastation — a creature card with mana value X or less",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					if x < 10 {
						return nil
					}
					fresh := NewContext(g, item)
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{
						Match:     yours,
						Power:     x,
						Toughness: x,
						Label:     "Finale of Devastation — +X/+X",
					}).Apply(fresh); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Match:    yours,
						Keywords: []string{"haste"},
						Label:    "Finale of Devastation — haste",
					}.Apply(fresh)
				},
			}.Apply(ctx)
		},
	})
}
