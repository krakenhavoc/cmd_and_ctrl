package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squee's Revenge — Sorcery {1}{U}{R} (Apocalypse, S82 OLD, #5614):
//
//	"Choose a number. Flip a coin that many times or until you lose a flip,
//	 whichever comes first. If you win all the flips, draw two cards for each
//	 flip."
//
// The number is chosen as the spell resolves (CR 608.2d; it may be zero, which
// flips nothing and draws nothing). The flips are called one at a time (CR 705.2),
// each win asking the next, and the first loss ends the chain; winning the last
// of the chosen number draws two cards for each.
//
// No simplification.
func init() {
	registerOldB629SqueesRevengeChain()
	effects.Register(effects.Spec{
		OracleID:     "3c9f3f26-339e-459e-9847-4e33aafb6f9b",
		Name:         "Squee's Revenge",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseNumber{
				Min:      0,
				NoMax:    true,
				Question: "Squee's Revenge — choose a number",
				Unit:     game.PayAmountOther,
				Goal:     func(*effects.Context) int { return 1 },
				Then:     oldB629SqueesRevengeBegin,
			}.Apply(ctx)
		},
	})
}

// oldB629SqueesRevengeBegin is Squee's Revenge's answer: the chosen number of
// flips, one called flip at a time (CR 705.2), each won flip asking the next
// until the number is reached, and the first lost flip ending the chain. Only a
// chain that wins every flip draws, two cards for each.
var oldB629SqueesRevengeBegin = effects.NumberThen("number/s82-old-b629-squees-revenge",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		if r.Number <= 0 {
			return nil // every one of zero flips is won, and two cards for each is none
		}
		oldB629SqueesRevengeFlip(ctx.Game, ctx.Controller(), ctx.Source(), 0, r.Number)
		return nil
	})

// oldB629SqueesRevengeFlip queues one called flip of the chain; `wins` flips
// are already won and `total` is the number chosen.
func oldB629SqueesRevengeFlip(g *game.Game, flipper, source uuid.UUID, wins, total int) {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  flipper,
		Source:   source,
		Question: "Squee's Revenge — call the coin flip",
		Wins:     wins,
		ThenKey:  oldB629SqueesRevengeChained,
		Nums:     []int{total},
	})
}

// oldB629SqueesRevengeChained is a settled call. An on-disk key, registered in
// init because its body queues itself.
var oldB629SqueesRevengeChained game.CoinFlipThen

func registerOldB629SqueesRevengeChain() {
	oldB629SqueesRevengeChained = effects.CoinFlipThen("coin-flip/s82-old-b629-squees-revenge", func(ctx *effects.Context, r game.CoinFlipped) error {
		if !effects.WonFirstFlip(r) {
			return nil
		}
		total := effects.CarriedNum(r.Nums, 0)
		if wins := r.Wins + 1; wins < total {
			oldB629SqueesRevengeFlip(ctx.Game, r.Flipper, r.Source, wins, total)
			return nil
		}
		return effects.DrawCards{Player: r.Flipper, N: 2 * total}.Apply(ctx)
	})
}
