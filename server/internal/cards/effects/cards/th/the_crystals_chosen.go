package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Crystal's Chosen — Sorcery {5}{W}{W} (Final Fantasy, #4462):
//
//	"Create four 1/1 colorless Hero creature tokens. Then put a +1/+1
//	 counter on each creature you control."
//
// The tokens are made first, so the set of creatures read for the
// counters includes all four of them: each Hero ends up 2/2, and every
// other creature the caster controls as the spell resolves gets one
// counter too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df4ef2f8-2591-4be3-b905-4f9ff9c538ea",
		Name:         "The Crystal's Chosen",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard(effects.JobSelectHeroToken), N: 4}).Apply(ctx); err != nil {
				return err
			}
			return effects.PutACounterOnEachCreatureYouControl(game.CounterPlusOne)(ctx.Game, item)
		},
	})
}
