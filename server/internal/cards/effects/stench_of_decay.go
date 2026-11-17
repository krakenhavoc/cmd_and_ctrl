package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stench of Decay — Instant {1}{B}{B} (Alliances, #5634):
//
//	"Nonartifact creatures get -1/-1 until end of turn."
//
// A layer 7c change (CR 613.4c) to every nonartifact creature on the
// battlefield, whoever controls it, ending at cleanup (CR 514.2). The set
// is fixed as it resolves (CR 611.2c), so a creature that enters
// afterwards is not shrunk. It sweeps the X/1s, so it declares a -1/-1
// sweep that spares artifact creatures (ADR 0126 §6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d4fbb16-0421-4e9c-af9e-e2c12a11c4ed",
		Name:         "Stench of Decay",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match: And(Creature(), Not(Artifact())), Power: -1, Toughness: -1,
				Label: "Stench of Decay — -1/-1 until end of turn",
			}.Apply(ctx)
		},
	})
}
