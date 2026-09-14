package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abandon Attachments — Instant — Lesson {1}{U/R} (EDHREC rank 3697):
//
//	"You may discard a card. If you do, draw two cards."
//
// A two-mana loot-into-two, and a Lesson for the learn family (the
// type line is what "Lesson spell" watchers read, so nothing on the
// card is needed for that).
//
// Sandbox simplification, declared in full because it is the one
// that changes how the card plays — Victimize's posture one zone
// over: the discard is modelled as an ADDITIONAL COST TO CAST
// rather than as an optional resolution-time action. The engine's
// resolution-time discard prompt is a fixed count with no
// continuation and no "you may", so "you may discard a card. If you
// do, …" has no shape mid-resolution, and a discard cost is the only
// discard-with-a-choice shape that exists (Thrill of Possibility's).
// The observable differences all run the weaker way:
//
//   - The card is discarded at announce, so a discard payoff
//     triggers ABOVE the spell and resolves first, and countering
//     the spell does not give the card back.
//   - With no other card in hand the spell cannot be cast at all
//     (printed, it can be cast to do nothing).
//   - The discard is never declined: casting it means discarding.
//
// The draw is the printed two, on resolution.
func init() {
	Register(Spec{
		OracleID:       "82333385-631f-4abf-b159-bb367f1c6fd9",
		Name:           "Abandon Attachments",
		Completeness:   CompletenessCaveats,
		Caveats:        []string{"You must discard a card as you cast it rather than choosing on resolution, so you can't cast it with no other card in hand and countering it doesn't give the card back."},
		AdditionalCost: DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
