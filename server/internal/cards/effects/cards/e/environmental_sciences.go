package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2f889671-a1d6-4034-881b-58c93897d707",
		Name:         "Environmental Sciences",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:    ctx.Controller(),
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Environmental Sciences — a basic land card",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: 2}.Apply(ctx)
			}),
		),
	})
}
