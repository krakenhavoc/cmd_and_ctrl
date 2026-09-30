package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entreat the Angels — Sorcery {X}{X}{W}{W}{W} (#1665):
//
//	"Create X 4/4 white Angel creature tokens with flying.
//	 Miracle {X}{W}{W} (You may cast this card for its miracle cost
//	 when you draw it if it's the first card you drew this turn.)"
//
// The miracle cost carries its own X, so CR 107.3b does not lock X at
// zero: a miracle cast announces X exactly as a hard cast does, and
// pays it once rather than twice — five Angels for {5}{W}{W} instead
// of {10}{W}{W}{W}.
func init() {
	Register(Spec{
		OracleID:         "b349f018-c20b-48b0-9e65-d5fd56b24b88",
		Name:             "Entreat the Angels",
		Completeness:     CompletenessFull,
		XMatters:         true,
		AlternativeCosts: []game.AlternativeCost{Miracle("{X}{W}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := ctx.X()
			if n <= 0 {
				return nil
			}
			return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Angel with flying"), N: n}.Apply(ctx)
		},
	})
}
