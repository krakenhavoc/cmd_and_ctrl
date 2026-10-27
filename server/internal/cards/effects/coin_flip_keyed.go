package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// coin_flip_keyed.go — #4253 slice 10b: a called coin flip's "if you win
// the flip, …" as a registered continuation (game.RegisterCoinFlipThen)
// plus the plain data it reads, so a table waiting on the call is a
// restore point.
//
// A card registers its continuation once, from a package-level var, under
// an on-disk key ("coin-flip/<card>-<what>", ledgered "coinflip <key>":
// never renamed, never reused), and passes it as game.CoinFlipSpec.ThenKey.
// What the closure captured rides the prompt: the asking item
// (CoinFlipSpec.Asked, game.NewAskedItem) for a body that reads its
// targets, its source object or its trigger, IDs (Carry), numbers (Nums)
// and words (Words).

// CoinFlipThen registers a keyed coin-flip continuation whose body runs
// against a Context rebuilt from the asking item when the flip carried
// one, else a stand-in with the flipper as the controller and the asking
// card as the source.
func CoinFlipThen(key string, body func(ctx *Context, r game.CoinFlipped) error) game.CoinFlipThen {
	return game.RegisterCoinFlipThen(key, func(g *game.Game, r game.CoinFlipped) error {
		ctx, _ := askedContext(g, r.Asked, r.Flipper, r.Source, nil)
		return body(ctx, r)
	})
}

// wonFirstFlip reports whether the first coin of a settled call was won.
func wonFirstFlip(r game.CoinFlipped) bool {
	return len(r.Result.Won) > 0 && r.Result.Won[0]
}

// carriedID is the i-th carried ID, or uuid.Nil past the end.
func carriedID(carry []uuid.UUID, i int) uuid.UUID {
	if i < 0 || i >= len(carry) {
		return uuid.Nil
	}
	return carry[i]
}

// carriedNum is the i-th carried number, or 0 past the end.
func carriedNum(nums []int, i int) int {
	if i < 0 || i >= len(nums) {
		return 0
	}
	return nums[i]
}

// carriedWord is the i-th carried word, or "" past the end.
func carriedWord(words []string, i int) string {
	if i < 0 || i >= len(words) {
		return ""
	}
	return words[i]
}
