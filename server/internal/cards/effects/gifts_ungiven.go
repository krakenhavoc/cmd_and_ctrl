package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gifts Ungiven — Instant {3}{U}:
//
//	"Search your library for up to four cards with different names and
//	 reveal them. Target opponent chooses two of those cards. Put the
//	 chosen cards into your graveyard and the rest into your hand.
//	 Then shuffle."
//
// Intuition's sibling on the "opponent picks from a revealed set" seam
// row (#1214), and the card that needs BOTH set-level hooks:
//
//   - "with different names" is a rule about the SET the search
//     returns, which no per-card predicate can state, so it rides
//     SearchLibrarySpec.Validate (#1017 / #682's shape);
//   - the opponent's "chooses two" is a reveal_pick with floor and
//     ceiling two, clamped to what was actually found — a search that
//     found one card is a choice of one (CR 608.2's "as much as
//     possible").
//
// The chosen cards go to the GRAVEYARD and the rest to the hand, which
// is the inversion of Intuition's last sentence and the only difference
// between the two cards' continuations.
//
// # The shuffle is last, and that is not the printed order
//
// The card says "then shuffle" after the choice, and paper plays it by
// setting the found cards aside during the shuffle. The engine has no
// set-aside zone, so the cards stay in the library — revealed, and
// therefore known to everyone — until the opponent has answered, and
// the shuffle runs at the end. The final zones are identical either
// way, and doing it the printed way round would wipe the per-card
// knowledge the reveal just granted and hand the opponent an empty
// prompt. Not a caveat: nothing observable differs.
func init() {
	Register(Spec{
		OracleID:     "58aec411-167d-4709-8560-793eaaed62c5",
		Name:         "Gifts Ungiven",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return giftsUngivenResolve(item, ctx)
		},
	})
}

// giftsUngivenResolve searches for up to four differently-named cards
// and hands them to the target opponent.
//
// Caller holds g.mu.
func giftsUngivenResolve(item *game.StackItem, ctx *Context) error {
	opp := targetedOpponent(item, ctx.Game)
	if opp == uuid.Nil {
		return nil
	}
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:  ctx.Controller(),
		Source:  ctx.Source(),
		Dest:    game.ZoneLibrary,
		Limit:   4,
		Reveal:  true,
		Shuffle: false,
		// See Intuition: ToTop is what makes the search report what it
		// found, and the shuffle printed after the choice wipes the
		// order it imposes.
		ToTop:    true,
		Reason:   "Gifts Ungiven — search your library for up to four cards with different names",
		Validate: DifferentNames,
		Then:     giftsUngivenChoose(item, opp),
	})
}

// giftsUngivenChoose is the continuation of the search.
func giftsUngivenChoose(item *game.StackItem, opp uuid.UUID) func(g *game.Game, found []uuid.UUID) error {
	return func(g *game.Game, found []uuid.UUID) error {
		ctx := NewContext(g, item)
		if len(found) == 0 {
			return g.ShuffleLibraryForEffect(ctx.Controller())
		}
		return RevealPick{
			Player:   opp,
			Owner:    ctx.Controller(),
			Question: "Gifts Ungiven — choose two of these cards; they go to their graveyard and the rest to their hand",
			Cards:    found,
			Min:      2,
			Max:      2,
			Then:     giftsUngivenSettle,
		}.Apply(ctx)
	}
}

// giftsUngivenSettle is Intuition's last sentence inverted: the CHOSEN
// cards go to the graveyard.
//
// Caller holds g.mu.
func giftsUngivenSettle(ctx *Context, picked, left []uuid.UUID) error {
	return revealPickSettle(ctx, picked, left, false)
}

// DifferentNames is the set-level legality hook behind "cards with
// different names" (CR 701.19c's reading of a search clause): no two
// cards in the picked set share a name.
//
// Here rather than in the card file because it is the shape of the
// clause and not of the card — the same sentence is printed on several
// tutors — and because a set-level rule is exactly what
// SearchLibrarySpec.Validate and ChooseCardsPrompt.Validate exist for.
//
// Names are compared case-insensitively and with surrounding space
// trimmed, which matches every other name comparison in the catalog. A
// card with no name (a fixture) matches nothing and is allowed
// through: this hook is a restriction, not an identity check.
func DifferentNames(picked []game.Card) bool {
	seen := make(map[string]bool, len(picked))
	for _, c := range picked {
		key := strings.ToLower(strings.TrimSpace(c.Name))
		if key == "" {
			continue
		}
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}
