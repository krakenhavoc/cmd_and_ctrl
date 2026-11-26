package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roilmage's Trick — Instant {3}{U} (Battle for Zendikar):
//
//	"Converge — Creatures your opponents control get -X/-0 until end of turn,
//	 where X is the number of colors of mana spent to cast this spell.
//	 Draw a card."
//
// The creatures your opponents control as it resolves are the ones shrunk
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "73eaf40a-abc6-486a-958d-374ed6111173",
		Name:                "Roilmage's Trick",
		Completeness:        effects.CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Purpose:             game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -ctx.ColorsSpentCount(),
					Label: "Roilmage's Trick — -X/-0",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 1}.Apply(ctx) }),
		),
	})
}
