package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "dc600c06-8239-409e-b53d-20f813a3f5e7",
		Name:         "Grand Crescendo",
		Purpose:      game.Purpose{Answers: game.AnswerSurvive | game.AnswerMakesBlocker}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The indestructible grant waits for a prompt the tokens' creation asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				n := ctx.X()
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: n}.Apply(ctx)
			}),
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"indestructible"},
				Label:    "Grand Crescendo — creatures you control gain indestructible",
			},
		),
	})
}
