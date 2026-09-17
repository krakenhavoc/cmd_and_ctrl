package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirst for Knowledge — Instant {2}{U} (EDHREC rank 2026):
//
//	"Draw three cards. Then discard two cards unless you discard an
//	 artifact card."
//
// The artifact deck's instant-speed draw-three. The draw happens
// first, so the drawn cards are legal discards, exactly as in paper;
// the discard is the player's own choice through the discard modal.
//
// DECLARED SIMPLIFICATION: the "unless you discard an artifact card"
// branch is not offered. The engine's discard prompt is a plain
// count — it has no way to accept EITHER one artifact card OR two
// cards of any kind — so the spell always asks for two. That is the
// weaker half of the printed choice and never the stronger one: a
// player with an artifact in hand may still discard it, and one
// more card besides. #651 gave the discard prompt the set-level
// Validate hook the search prompt has (DiscardPrompt.Validate), so
// the branch is now expressible; offering it is card work this
// engine fix deliberately left alone.
func init() {
	Register(Spec{
		OracleID:     "939e6f71-185e-41f2-9d54-72cce06f1dce",
		Name:         "Thirst for Knowledge",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You always discard two cards — discarding a single artifact card instead isn't offered."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b16DrawThenDiscard(ctx.Game, item, 3, 2)
		},
	})
}
