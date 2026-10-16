package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strange Inversion — Instant — Arcane {2}{R}:
//
//	"Switch target creature's power and toughness until end of turn.
//	Splice onto Arcane {1}{R}"
func init() {
	Register(Spec{
		OracleID:     "540f8779-cf29-4582-8b7f-95ebcae92925",
		Name:         "Strange Inversion",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{R}"),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return switchEachTargetPT(ctx, "Strange Inversion — switch power and toughness")
		},
	})
}
