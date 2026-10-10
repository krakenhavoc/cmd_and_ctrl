package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grand Crescendo — Instant {X}{W}{W}:
//
//	"Create X 1/1 green and white Citizen creature tokens. Creatures
//	 you control gain indestructible until end of turn."
//
// The grant is evaluated after the tokens exist, so the new Citizens
// are among the creatures that gain indestructible (CR 611.2c reads
// the set when the instruction is carried out, and "creatures you
// control" at that moment includes them).
//
// The indestructible grant is a fixed rider that happens whatever X
// is, so X=0 is a real move and the card leaves XMatters unset.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc600c06-8239-409e-b53d-20f813a3f5e7",
		Name:         "Grand Crescendo",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if n := ctx.X(); n > 0 {
				if err := (CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"indestructible"},
				Label:    "Grand Crescendo — creatures you control gain indestructible",
			}.Apply(ctx)
		},
	})
}
