package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardians' Pledge — Instant {1}{W}{W} (Magic 2012, #5573):
//
//	"White creatures you control get +2/+2 until end of turn."
//
// Righteous Charge's team pump limited to your white creatures: the set is
// fixed as it resolves (CR 611.2c) and the bonus ends in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32d78ffd-5bed-45e8-be6a-420149a263bb",
		Name:         "Guardians' Pledge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl(), effects.OfColor("W")), Power: 2, Toughness: 2,
				Label: "Guardians' Pledge — white creatures you control get +2/+2",
			}.Apply(ctx)
		},
	})
}
