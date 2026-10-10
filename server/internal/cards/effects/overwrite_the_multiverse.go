package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overwrite the Multiverse — Sorcery {4}{B}{B} (Reality Fracture,
// tracker #2795):
//
//	"Exile all creatures. Empower Jace X, where X is the number of
//	 creatures exiled this way."
//
// ExileAllMatching's Then runs from a continuation once every creature
// has finished moving (a commander's CR 903.9 prompt holds it), with the
// count of cards that really reached exile; the keyword action (ADR
// 0139) is then given that count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0372cde-bf30-4a1b-93bb-9ff4ac1c116d",
		Name:         "Overwrite the Multiverse",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ExileAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, _ []game.Card, exiled int) error {
					return EmpowerJace{N: exiled}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
