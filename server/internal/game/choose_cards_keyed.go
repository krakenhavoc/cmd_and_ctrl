package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// choose_cards_keyed.go — a choose_cards pick whose continuation is a
// KEY rather than a closure (#2040), so a table waiting on the answer is
// still a restore point.
//
// The option pick got this in #2854 (option_pick_keyed.go) and this is
// the same shape for the card-set pick: the continuation is a
// package-level function registered under an on-disk key, and what it
// needs travels on the prompt as plain data — the chooser, the card that
// asked, the candidates and bounds the prompt already carries, and a
// list of IDs the card put there (Carry). On the prompt the key is
// PendingChoice.PickThen and the carried IDs OptionCarry, the two fields
// the keyed option pick already writes, so a binary from before this
// change refuses a restore point naming a key it does not have
// (checkEffectKeys) rather than restoring a question whose answer would
// run nothing. The key shares the effect-key namespace and ledger
// ("choose <key>" in testdata/effect_keys.txt).
//
// What a keyed pick gives up is the closure-only half of
// ChooseCardsPrompt: the per-set Validate and the live-zone re-check
// (Zone). Both are closures or ride the closure frame. A continuation
// that cares where a pick is now checks it itself, as Immortal Coil's
// does.

// CardsChosen is what a keyed choose_cards continuation is handed.
type CardsChosen struct {
	// Chooser answered; Source is the card that asked.
	Chooser, Source uuid.UUID
	// Picked are the cards chosen, in the order submitted. Nil when the
	// prompt ended unanswered (its chooser left the game).
	Picked []uuid.UUID
	// Answered is false when the prompt ended unanswered.
	Answered bool
	// Candidates, Min and Max are the prompt as it was asked.
	Candidates []uuid.UUID
	Min, Max   int
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
}

// ChooseCardsFunc is a registered continuation. It runs with g.mu held,
// may queue prompts of its own, and must capture nothing.
type ChooseCardsFunc func(g *Game, r CardsChosen) error

// ChooseCardsThen names a registered continuation. The key is
// unexported, so the only way to hold one is RegisterChooseCardsThen.
type ChooseCardsThen struct{ key string }

// Key is the continuation's on-disk key.
func (r ChooseCardsThen) Key() string { return r.key }

// chooseCardsThens is guarded by effectRegistryMu, with the other
// effect keys, because the keys share one namespace and one ledger.
var chooseCardsThens = map[string]ChooseCardsFunc{}

// RegisterChooseCardsThen registers a keyed choose_cards continuation.
// Call it once, from a package-level var. Panics on a bad or duplicate
// key, as DelayedBody does. The key is an on-disk identity: never
// renamed, never reused.
func RegisterChooseCardsThen(key string, fn ChooseCardsFunc) ChooseCardsThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("choose-cards continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: choose-cards continuation %q has no function", key))
	}
	chooseCardsThens[key] = fn
	return ChooseCardsThen{key: key}
}

func lookupChooseCardsThen(key string) (ChooseCardsFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := chooseCardsThens[key]
	return fn, ok
}

// KnownChooseCardsThen reports whether this binary has a continuation
// registered under key.
func KnownChooseCardsThen(key string) bool { _, ok := lookupChooseCardsThen(key); return ok }

// registeredChooseCardsKeysLocked lists the ledger lines, "choose
// <key>". Caller holds effectRegistryMu.
func registeredChooseCardsKeysLocked() []string {
	out := make([]string, 0, len(chooseCardsThens))
	for k := range chooseCardsThens {
		out = append(out, "choose "+k)
	}
	sort.Strings(out)
	return out
}

// isKeyedChooseCards reports whether c is a choose_cards pick with a
// registered continuation.
func isKeyedChooseCards(c *PendingChoice) bool {
	return c != nil && c.Kind == PendingChoiceChooseCards && c.PickThen != "" && c.chooseCardsResume == nil
}

// runKeyedChooseCardsLocked runs c's continuation with `picked`
// (answered false: nobody chose). Caller must hold g.mu (write).
func (g *Game) runKeyedChooseCardsLocked(c *PendingChoice, picked []uuid.UUID, answered bool) error {
	fn, ok := lookupChooseCardsThen(c.PickThen)
	if !ok {
		return fmt.Errorf("%w: choose-cards continuation %q", ErrUnknownEffectKey, c.PickThen)
	}
	return fn(g, CardsChosen{
		Chooser:    c.Chooser,
		Source:     c.Source,
		Picked:     copyUUIDs(picked),
		Answered:   answered,
		Candidates: copyUUIDs(c.ChooseCards),
		Min:        c.ChooseMin,
		Max:        c.ChooseMax,
		Carry:      copyUUIDs(c.OptionCarry),
	})
}
