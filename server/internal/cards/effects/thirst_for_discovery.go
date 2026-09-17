package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirst for Discovery — Instant {2}{U} (EDHREC rank 3748):
//
//	"Draw three cards. Then discard two cards unless you discard a
//	 basic land card."
//
// Thirst for Knowledge with basic lands where that card has
// artifacts: three cards now, and the discard is either one basic
// land or any two. The draws land first, then the discard prompt
// opens over the whole hand, so the drawn cards are among the
// choices.
//
// Sandbox simplification, declared, Thirst for Knowledge's exactly:
// the "unless" is NOT offered. The card asks for a fixed count, so
// it cannot accept "one card, provided it is a basic land" as an
// alternative to "two cards" — the player always discards two.
// Weaker than printed (the cheaper discard is never available),
// never stronger. #651 gave the discard prompt the set-level
// Validate hook the search prompt has (DiscardPrompt.Validate), so
// the branch is now expressible; offering it is card work this
// engine fix deliberately left alone.
func init() {
	Register(Spec{
		OracleID:     "1e05e6ef-14af-451d-9d54-e75b1f8871ab",
		Name:         "Thirst for Discovery",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You always discard two cards — discarding a single basic land card instead isn't offered."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b16DrawThenDiscard(ctx.Game, item, 3, 2)
		},
	})
}
