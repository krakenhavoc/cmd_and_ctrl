package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// United Front — Sorcery {X}{W}{W} (Avatar: The Last Airbender, #4459):
//
//	"Create X 1/1 white Ally creature tokens, then put a +1/+1 counter on
//	 each creature you control."
//
// The tokens come first, so they are among "each creature you control"
// when the counters are put (Abandoned Air Temple's one placement per
// creature, read after the tokens exist). X is the value announced for
// the spell (CR 107.3). At X=0 there are no tokens and the counters still
// go on the creatures you already control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c23f260b-b7d3-4c25-a08a-dabdb99e1791",
		Name:         "United Front",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if x := ctx.X(); x > 0 {
				if err := (CreateToken{Template: TokenCard("1/1 white Ally"), N: x}).Apply(ctx); err != nil {
					return err
				}
			}
			return b11PutCounterOnEachCreatureYouControl(ctx.Game, item)
		},
	})
}
