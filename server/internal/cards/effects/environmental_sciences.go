package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Environmental Sciences — Sorcery — Lesson {2} (Mystery Booster 2, #5948):
//
//	"Search your library for a basic land card, reveal it, put it into your
//	 hand, then shuffle. You gain 2 life."
//
// The basic-land tutor to hand, then the life gain, in the printed order
// (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f889671-a1d6-4034-881b-58c93897d707",
		Name:         "Environmental Sciences",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:    ctx.Controller(),
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Environmental Sciences — a basic land card",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Amount: 2}.Apply(ctx)
			}),
		),
	})
}
