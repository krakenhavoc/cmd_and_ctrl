package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruel Tutor — Sorcery {2}{B} (Avatar: The Last Airbender Eternal, #5075):
//
//	"Search your library for a card, then shuffle and put that card on
//	 top. You lose 2 life."
//
// Vampiric Tutor's body at sorcery speed: the search shuffles, the
// found card goes on top of the library unrevealed (SearchLibrary.ToTop
// places it after the shuffle), and the life is lost after the search
// (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9353eb6-c7e7-4d85-a0c2-32b1d091881a",
		Name:         "Cruel Tutor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:  ctx.Controller(),
					Dest:    game.ZoneLibrary,
					ToTop:   true,
					Limit:   1,
					Shuffle: true,
					Reason:  "Cruel Tutor — search your library for a card",
				}.Apply(ctx)
			}),
			GainLife{Amount: -2},
		),
	})
}
