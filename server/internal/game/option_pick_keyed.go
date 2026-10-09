package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// option_pick_keyed.go — an option pick whose continuation is a KEY
// rather than a closure (#2854), so a table waiting on the answer is
// still a restore point.
//
// An ordinary option pick holds its continuation as a closure
// (optionPickFrame), and a func cannot be written to a restore point:
// while one is open, the census counts it (ChoiceResumeFrames) and the
// server keeps the last restore point from before the question. That
// is the right trade for a pile split, which is over in one answer.
// It is the wrong one for a question asked once per creature or per
// player — Winter's Chill asks every targeted creature's controller in
// turn, Lim-Dûl's Hex every player — where a restart in the middle
// would rewind the whole spell.
//
// So, as RegisterRevealedPickThen did for the revealed-hand pick
// (#2115), the continuation is a package-level function registered
// under an on-disk key, and everything it needs travels on the prompt
// as plain data: the chooser, the card that asked, the option chosen
// and a list of IDs the card put there (OptionPickPrompt.Carry — the
// creatures still to ask about, the spell to counter). The key shares
// the effect-key namespace and ledger (testdata/effect_keys.txt lists
// it as "option <key>"), and a restore point naming a key this binary
// does not have is refused and kept (checkEffectKeys), exactly like a
// body or a pick continuation.
//
// On the prompt the key is PendingChoice.PickThen, the field the
// revealed-hand pick's key already uses, and the carried IDs are
// OptionCarry. Reusing PickThen is what makes this additive within v7:
// every binary since #2115 refuses a file whose pickThen it cannot
// find, so a binary from before this change refuses a restore point
// with a keyed option pick open rather than restoring a question whose
// answer would run nothing.

// OptionPicked is what a keyed option pick's continuation is handed.
type OptionPicked struct {
	// Chooser answered; Source is the card that asked.
	Chooser, Source uuid.UUID
	// Index is the offset of the chosen option in the list the chooser
	// was SHOWN, or NoChoiceIndex when the prompt ended unanswered (its
	// chooser left the game). Options that cost mana the chooser could
	// not pay were never shown, so read Option, not Index, to learn
	// which branch was taken.
	Index int
	// Option is the chosen option, nil when Index is NoChoiceIndex. Its
	// ManaCost has already been paid.
	Option *ChoiceOption
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
}

// OptionPickFunc is a registered continuation. It runs with g.mu held,
// may queue prompts of its own, and must capture nothing.
type OptionPickFunc func(g *Game, r OptionPicked) error

// OptionPickThen names a registered continuation. The key is
// unexported, so the only way to hold one is RegisterOptionPickThen.
type OptionPickThen struct{ key string }

// Key is the continuation's on-disk key.
func (r OptionPickThen) Key() string { return r.key }

// optionPickThens is guarded by effectRegistryMu, with the other
// effect keys, because the keys share one namespace and one ledger.
var optionPickThens = map[string]OptionPickFunc{}

// RegisterOptionPickThen registers a keyed option-pick continuation.
// Call it once, from a package-level var. Panics on a bad or duplicate
// key, as DelayedBody does. The key is an on-disk identity: never
// renamed, never reused.
func RegisterOptionPickThen(key string, fn OptionPickFunc) OptionPickThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("option pick continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: option pick continuation %q has no function", key))
	}
	optionPickThens[key] = fn
	return OptionPickThen{key: key}
}

func lookupOptionPickThen(key string) (OptionPickFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := optionPickThens[key]
	return fn, ok
}

// KnownOptionPickThen reports whether this binary has a continuation
// registered under key.
func KnownOptionPickThen(key string) bool { _, ok := lookupOptionPickThen(key); return ok }

// registeredOptionPickKeysLocked lists the ledger lines, "option
// <key>". Caller holds effectRegistryMu.
func registeredOptionPickKeysLocked() []string {
	out := make([]string, 0, len(optionPickThens))
	for k := range optionPickThens {
		out = append(out, "option "+k)
	}
	sort.Strings(out)
	return out
}

// RunOptionPickThenForEffect runs a keyed continuation directly, for
// the caller whose prompt could not be queued (QueueOptionPickForEffect
// returned uuid.Nil: its chooser has left the game) and which still
// has the rest of the card to run, as PickOption's closure form does
// with -1. Caller must hold g.mu.
func (g *Game) RunOptionPickThenForEffect(then OptionPickThen, r OptionPicked) error {
	if then.key == "" {
		return nil
	}
	return runOptionPickThen(g, then.key, r)
}

// runOptionPickThen runs the continuation registered under key.
// Caller must hold g.mu.
func runOptionPickThen(g *Game, key string, r OptionPicked) error {
	fn, ok := lookupOptionPickThen(key)
	if !ok {
		return fmt.Errorf("%w: option pick continuation %q", ErrUnknownEffectKey, key)
	}
	return fn(g, r)
}
