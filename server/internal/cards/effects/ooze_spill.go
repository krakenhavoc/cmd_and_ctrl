package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ooze Spill — {1}{U}{U} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Counter target spell. Create a Mutagen token. (It's an artifact with
//	 "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// Absorb with a token for the rider: the spell is countered if it still
// can be, and the token is made either way, so a spell that can't be
// countered still pays it out. With the target gone before resolution the
// whole spell does not resolve (CR 608.2b), and no token is made.
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1c5f9da-1396-4406-b803-b64f66b8e49d",
		Name:         "Ooze Spill",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: a counterspell (the enumerator reads the counter itself; remove, as sweep ruling 8's stack removal)
		Targets:      TargetSpell("target spell"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return CounterTarget{StackID: t.ID}.Apply(ctx)
					}
				}
				return nil
			}),
			CreateToken{Template: MutagenToken(), N: 1},
		),
	})
}
