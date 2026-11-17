package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terisiare's Devastation — Sorcery {X}{2}{B}{B} (The Brothers' War):
//
//	"You lose X life and create X tapped Powerstone tokens. Then all
//	 creatures get -1/-1 until end of turn for each artifact you control."
//
// Three instructions in order. The artifacts are counted after the
// tokens exist ("then"), so the new Powerstones are in the count, and
// the count is made once as the effect resolves (CR 611.2c) and applies
// to every creature on the battlefield then, the caster's own included.
// The life loss may ask a question (a replacement ordering), so each
// instruction is a step that waits for the one before it (CR 608.2c).
// With X = 0 nothing is lost or created, and the shrink still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03c1e376-1410-4321-90fe-a6997e8db85a",
		Name:         "Terisiare's Devastation",
		Completeness: CompletenessFull,
		XMatters:     true,
		// The shrink is -1/-1 per artifact, and the X Powerstones it makes are
		// artifacts you control, so it is at least -X/-X: declared as X.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		OnResolve: ResolveSteps(
			Instruction(loseXLife),
			Instruction(func(ctx *Context) error {
				if x := ctx.X(); x > 0 {
					return CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: x}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				n := b03ArtifactsControlled(ctx.Game, ctx.Controller())
				if n <= 0 {
					return nil
				}
				return BoostUntilEOT{Match: Creature(), Power: -n, Toughness: -n, Label: "Terisiare's Devastation — -1/-1 for each artifact you control"}.Apply(ctx)
			}),
		),
	})
}
