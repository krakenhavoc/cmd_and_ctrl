package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crop Rotation — Instant {G}:
//
//	"As an additional cost to cast this spell, sacrifice a land.
//	 Search your library for a land card, put that card onto the
//	 battlefield, then shuffle."
//
// One green mana turns a land that has done its job into any land in
// the deck, at instant speed. The batch-01 triage filed it under
// "cost modification"; Harrow's file already records why that is the
// wrong drawer. "As an additional cost, sacrifice a land" is CR
// 601.2f — an ADDITIONAL cost, which `Spec.AdditionalCost` has
// carried since the S21 pass — and nothing about the mana cost is
// modified.
//
// Paying the cost with the spell already on the stack is the point of
// doing it that way (CR 601.2a before 601.2h): the sacrificed land's
// dies-trigger goes ABOVE Crop Rotation and resolves first, and a
// countered Crop Rotation does not hand the land back.
//
// Any land, not just a basic — the searched card enters untapped,
// because the printed text has no "tapped" clause, and the fetched
// land's own enters-tapped replacement still runs (#263).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "28b46183-c62f-47b1-9fee-3ba148202cab",
		Name:           "Crop Rotation",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsLand() },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Crop Rotation — a land card",
			}.Apply(ctx)
		},
	})
}
