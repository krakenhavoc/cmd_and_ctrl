package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Puppet's Verdict — Instant {1}{R}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Flip a coin. If you win the flip, destroy all creatures with power 2 or
//	 less. If you lose the flip, destroy all creatures with power 3 or greater."
//
// Boompile's keyed flip: the caster calls it as the spell resolves, and the winner
// picks which half of the creatures dies. Power is read as the destruction happens
// (a pump that resolved in response counts), and every creature on the battlefield
// is a candidate, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e07a7c75-b7b7-4c2d-bd7d-5f056cde871d",
		Name:         "Puppet's Verdict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Puppet's Verdict — call the coin flip",
				ThenKey:  puppetsVerdictFlipped,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// puppetsVerdictFlipped is the flip's answer. An on-disk key.
var puppetsVerdictFlipped = effects.CoinFlipThen("coin-flip/puppets-verdict", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return effects.DestroyAllMatching{Match: func(g *game.Game, you uuid.UUID, c game.Card) bool {
			return effects.Creature()(g, you, c) && effects.PowerLE(2)(g, you, c)
		}}.Apply(ctx)
	}
	return effects.DestroyAllMatching{Match: func(g *game.Game, you uuid.UUID, c game.Card) bool {
		return effects.Creature()(g, you, c) && effects.PowerGE(3)(g, you, c)
	}}.Apply(ctx)
})
