package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirst for Knowledge — Instant {2}{U} (EDHREC rank 2026):
//
//	"Draw three cards. Then discard two cards unless you discard an
//	 artifact card."
//
// The artifact deck's instant-speed draw-three. The draw happens
// first, so the drawn cards are legal discards, exactly as in paper;
// the discard is the player's own choice through the discard modal,
// which accepts either two cards of any kind or a single artifact card
// (discardTwoUnlessOne, the Compulsive Research shape: a floor of one
// and a rule on the set).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "939e6f71-185e-41f2-9d54-72cce06f1dce",
		Name:         "Thirst for Knowledge",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: drawThenDiscardTwoUnlessArtifact, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}
