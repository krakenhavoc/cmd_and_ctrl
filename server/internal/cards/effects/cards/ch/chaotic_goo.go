package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaotic Goo — Creature — Ooze {2}{R}{R}, 0/0 (Secret Lair Drop):
//
//	"This creature enters with three +1/+1 counters on it.
//	 At the beginning of your upkeep, you may flip a coin. If you win the
//	 flip, put a +1/+1 counter on this creature. If you lose the flip, remove
//	 a +1/+1 counter from this creature."
//
// The flip is the engine's called flip (Boompile's); the counter lands on or
// leaves this object only if it is still on the battlefield when the flip
// settles.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1005625-0a78-4f8e-bf51-80458fc0e9d6",
		Name:         "Chaotic Goo",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(game.CounterPlusOne, 3, "This creature enters with three +1/+1 counters on it")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Chaotic Goo — flip a coin: win, put a +1/+1 counter on it; lose, remove one",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Chaotic Goo — call the coin flip",
						ThenKey:  chaoticGooFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}), "Chaotic Goo — flip a coin?"),
		},
	})
}

var chaoticGooFlipped = effects.CoinFlipThen("coin-flip/chaotic-goo", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.OnBattlefield(ctx.Game, r.Source) {
		return nil
	}
	if effects.WonFirstFlip(r) {
		return effects.AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	if c, ok := ctx.Game.LookupCardForEffect(r.Source); ok && c.Counters[game.CounterPlusOne] > 0 {
		return effects.AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: -1}.Apply(ctx)
	}
	return nil
})
