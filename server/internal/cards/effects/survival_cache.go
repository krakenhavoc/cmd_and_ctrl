package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Survival Cache — Sorcery {2}{W}:
//
//	"You gain 2 life. Then if you have more life than an opponent, draw
//	 a card.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// "More life than an opponent" is more than at least one opponent
// still in the game, checked after the life gain. Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5fb8be5a-3666-4680-84e2-341cb269df07",
		Name:            "Survival Cache",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			// The comparison waits for the gain: a life change can pause
			// on a CR 616 ordering prompt, and "then" reads the totals
			// after it.
			return ctx.Game.ChangePlayerLifeThenForEffect(ctx.Source(), me, 2, func(g *game.Game, _ int) error {
				return survivalCacheDrawIfAhead(NewContext(g, item), me)
			})
		},
	})
}

// survivalCacheDrawIfAhead is "if you have more life than an opponent,
// draw a card": more than at least one opponent still in the game.
func survivalCacheDrawIfAhead(ctx *Context, me uuid.UUID) error {
	mine := ctx.PlayerByID(me)
	if mine == nil {
		return nil
	}
	for _, opp := range ctx.Opponents() {
		if p := ctx.PlayerByID(opp); p != nil && mine.Life > p.Life {
			return DrawCards{Player: me, N: 1}.Apply(ctx)
		}
	}
	return nil
}
