package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hinder — Instant {1}{U}{U}:
//
//	"Counter target spell. If that spell is countered this way, put
//	 that card on your choice of the top or bottom of its owner's
//	 library instead of into that player's graveyard."
//
// #1298, on ADR 0088's put_in_library (2026-09-23 amendment). The
// choice is Hinder's CONTROLLER's — "your choice" — and the card goes
// to its OWNER's library, so the chooser and the owner are usually two
// different players. That is why this waited: the prompt had to be
// able to ask one player about another's card, and the counter had to
// be able to take a position.
//
// The choice comes before the counter, in the order the card resolves:
// a `top_or_bottom` prompt (which always asks, even about one card),
// then CounterSpellToLibraryThenForEffect to the chosen end. So the
// spell is still on the stack while its fate is decided, a commander's
// owner is offered the command zone (CR 903.9) knowing where it was
// headed, flashback's exile wins, and a spell that can't be countered
// is neither asked about nor moved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9db6b94-a7b1-4b93-b454-4dead8f85e34",
		Name:         "Hinder",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterToLibrary{
				StackID:   item.Targets[0].ID,
				Placement: game.LibraryPlaceTopOrBottom,
				Label:     "Hinder — put that card on the top or bottom of its owner's library",
			}.Apply(ctx)
		},
	})
}
