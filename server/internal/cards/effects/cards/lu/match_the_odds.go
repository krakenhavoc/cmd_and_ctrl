package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Match the Odds — Sorcery — Lesson {2}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Create a 1/1 white Ally creature token. Put a +1/+1 counter on it
//	 for each creature your opponents control."
//
// Iroh, Tea Master's Ally body: the token is made, then the creatures
// your opponents control are counted as the spell resolves, through the
// layers (CR 608.2h), and that many +1/+1 counters go on the token as
// one placement. With none, the Ally stays a 1/1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "797d8776-9dcf-45b1-8c5e-1d025f916773",
		Name:         "Match the Odds",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			made, err := g.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("1/1 white Ally")), 1, game.TokenEntryOptions{})
			if err != nil || len(made) == 0 {
				return err
			}
			g.RecomputeLayersIfStaleLocked()
			n := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller != ctx.Controller() {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return effects.AddCounter{Target: made[0], Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		},
	})
}
