package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// coin_flip_keyed.go — a called coin flip whose continuation is a KEY
// rather than a closure (#4253 slice 10b), so a table waiting on the call is
// a restore point.
//
// A coin call (PendingChoiceCoinCall) held the rest of the card as a Go
// closure on its coinFlipResume frame, and a prompt with a frame is not
// written to a restore point. The keyed form carries no frame: the call's
// own data (flipper, source, coins, the stop rule, the wins so far) is
// already on the prompt, and the continuation is a package-level function
// registered under an on-disk key (RegisterCoinFlipThen), handed the result
// and the plain data the card put on the prompt: IDs (Carry, in
// PendingChoice.OptionCarry), numbers (Nums, PickNums), words (Words,
// PickWords) and the asking stack item (Asked).
//
// A flip with no continuation at all (FlipACoin: what winning does is
// whatever triggers on the EventFlipCoin) is keyed too, under the engine's
// own coinFlipNothing, so it needs no frame either.
//
// On the prompt the key is PendingChoice.PickThen, the field every keyed
// prompt shares, so a binary from before this change refuses a restore
// point naming a key it does not have (checkEffectKeys). The key shares
// the effect-key namespace and the ledger ("coinflip <key>").

// CoinFlipped is what a keyed coin-flip continuation is handed.
type CoinFlipped struct {
	// Flipper called the coin; Source is the card that asked.
	Flipper, Source uuid.UUID
	// Result is the call and the coins, as a closure's Then is handed it.
	Result CoinFlipResult
	// Coins, AllowStop, MaxUsefulWins and Wins are the call as it was
	// asked, for a continuation that asks the next link of a chain.
	Coins         int
	AllowStop     bool
	MaxUsefulWins int
	Wins          int
	// Carry, Nums and Words are the plain data the queuing card put on the
	// prompt (CoinFlipSpec.Carry, Nums, Words).
	Carry []uuid.UUID
	Nums  []int
	Words []string
	// Asked is a copy of the asking stack item when the card put one on
	// the prompt (CoinFlipSpec.Asked); nil otherwise.
	Asked *StackItem
}

// CoinFlipFunc is a registered continuation. It runs with g.mu held, may
// queue prompts of its own (the next flip of a chain), and must capture
// nothing.
type CoinFlipFunc func(g *Game, r CoinFlipped) error

// CoinFlipThen names a registered continuation. The key is unexported, so
// the only way to hold one is RegisterCoinFlipThen.
type CoinFlipThen struct{ key string }

// Key is the continuation's on-disk key.
func (r CoinFlipThen) Key() string { return r.key }

// coinFlipThens is guarded by effectRegistryMu, with the other effect keys,
// because the keys share one namespace and one ledger.
var coinFlipThens = map[string]CoinFlipFunc{}

// RegisterCoinFlipThen registers a keyed coin-flip continuation. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The key
// is an on-disk identity: never renamed, never reused.
func RegisterCoinFlipThen(key string, fn CoinFlipFunc) CoinFlipThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("coin-flip continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: coin-flip continuation %q has no function", key))
	}
	coinFlipThens[key] = fn
	return CoinFlipThen{key: key}
}

func lookupCoinFlipThen(key string) (CoinFlipFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := coinFlipThens[key]
	return fn, ok
}

// KnownCoinFlipThen reports whether this binary has a continuation
// registered under key.
func KnownCoinFlipThen(key string) bool { _, ok := lookupCoinFlipThen(key); return ok }

// registeredCoinFlipKeysLocked lists the ledger lines, "coinflip <key>".
// Caller holds effectRegistryMu.
func registeredCoinFlipKeysLocked() []string {
	out := make([]string, 0, len(coinFlipThens))
	for k := range coinFlipThens {
		out = append(out, "coinflip "+k)
	}
	sort.Strings(out)
	return out
}

// coinFlipNothing is a called flip with nothing hanging off it (FlipACoin):
// the flip's EventFlipCoin is the whole consequence. An on-disk key.
var coinFlipNothing = RegisterCoinFlipThen("coin-flip/nothing", func(*Game, CoinFlipped) error { return nil })

// keyedCoinFlip writes a keyed flip's continuation and data onto the
// prompt about to be queued: the spec's key, or coinFlipNothing for a spec
// with no continuation at all.
func (spec CoinFlipSpec) keyedCoinFlip(c *PendingChoice) {
	key := spec.ThenKey.key
	if key == "" {
		key = coinFlipNothing.key
	}
	c.PickThen = key
	c.OptionCarry = copyUUIDs(spec.Carry)
	c.PickNums = copyInts(spec.Nums)
	c.PickWords = copyStrings(spec.Words)
	c.Asked = copyAskedItem(spec.Asked)
}

// coinFlipSpecOf is the spec a coin call was asked with: its closure
// frame's, or for a keyed call the prompt's own data.
func coinFlipSpecOf(c *PendingChoice) (CoinFlipSpec, bool) {
	if c.coinFlipResume != nil {
		return c.coinFlipResume.spec, true
	}
	if c.PickThen == "" {
		return CoinFlipSpec{}, false
	}
	return CoinFlipSpec{
		Flipper: c.FromPlayer, Source: c.Source, Question: c.Reason, Coins: c.CoinCount,
		AllowStop: c.CoinAllowStop, MaxUsefulWins: c.CoinMaxUsefulWins, Wins: c.CoinWins,
		ThenKey: CoinFlipThen{key: c.PickThen}, Carry: copyUUIDs(c.OptionCarry),
		Nums: copyInts(c.PickNums), Words: copyStrings(c.PickWords), Asked: copyAskedItem(c.Asked),
	}, true
}

// runCoinFlipThen runs the keyed continuation of a settled flip. A key this
// binary lacks is an error (checkEffectKeys refuses such a restore point,
// so this is the live-table guard).
//
// Caller must hold g.mu.
func (g *Game) runCoinFlipThen(spec CoinFlipSpec, result CoinFlipResult) error {
	fn, ok := lookupCoinFlipThen(spec.ThenKey.key)
	if !ok {
		return fmt.Errorf("%w: coin-flip continuation %q", ErrUnknownEffectKey, spec.ThenKey.key)
	}
	return fn(g, CoinFlipped{
		Flipper: spec.Flipper, Source: spec.Source, Result: result,
		Coins: spec.Coins, AllowStop: spec.AllowStop, MaxUsefulWins: spec.MaxUsefulWins, Wins: spec.Wins,
		Carry: copyUUIDs(spec.Carry), Nums: copyInts(spec.Nums), Words: copyStrings(spec.Words),
		Asked: spec.Asked.StackItem(),
	})
}
