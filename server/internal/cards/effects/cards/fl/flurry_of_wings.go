package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flurry of Wings — Instant {G}{W}{U} (Alara Reborn, #5586):
//
//	"Create X 1/1 white Bird Soldier creature tokens with flying, where X
//	 is the number of attacking creatures."
//
// Galadhrim Ambush's count: every attacking creature, whoever controls it,
// read as the spell resolves (CR 608.2h). Outside combat X is 0 and
// nothing is created.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "259b3d11-02e2-43c6-9eed-4ffabb6f9fcf",
		Name:         "Flurry of Wings",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := len(effects.AttackingCreatures(ctx.Game))
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Bird Soldier with flying"), N: n}.Apply(ctx)
		},
	})
}
