package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "32d78ffd-5bed-45e8-be6a-420149a263bb",
		Name:         "Guardians' Pledge",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl(), OfColor("W")), Power: 2, Toughness: 2,
				Label: "Guardians' Pledge — white creatures you control get +2/+2",
			}.Apply(ctx)
		},
	})
}
