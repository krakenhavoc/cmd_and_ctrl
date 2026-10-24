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
// ChooseCardsPrompt: the live-zone re-check (Zone), which rides the
// closure frame. A continuation that cares where a pick is now checks it
// itself, as Immortal Coil's does.
//
// A set rule (ChooseCardsPrompt.Validate, "exactly two cards with
// different names") is registered under the same key as the
// continuation (RegisterChooseCardsThenRule, #4264) and looked up by it
// each time an answer is checked, so a keyed pick keeps its set rule and
// still carries no closure. And one word of the card's text, which a
// text-changing effect may have changed (CR 612.2), rides the prompt as
// CarryWord, in PendingChoice.PickMatchArg (#4264), for a continuation
// that asks the next prompt with it (Coax from the Blind Eternities).

// CardsChosen is what a keyed choose_cards continuation is handed.
type CardsChosen struct {
	// Chooser answered; Source is the card that asked. From is whose cards
	// they are (PendingChoice.FromPlayer): the owner of a reveal pick's
	// pool.
	Chooser, Source, From uuid.UUID
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
	// Question is the prompt's header as it was asked: a reveal names
	// it as its reason (#4256's outside-the-game picks).
	Question string
	// Word is the one word of the card's text the queuing card put on
	// the prompt (ChooseCardsPrompt.CarryWord, #4264). Empty when none.
	Word string
	// Key is the continuation's own key, as it rode the prompt.
	Key string
	// Offer and Asked are the any-player offer and the asking item a leg
	// of a keyed per-object offer carries (#4264); nil on every other
	// pick.
	Offer *AnyPlayerOffer
	Asked *StackItem
	// Run is the prompted run this pick is one leg of
	// (PendingChoice.promptRun, #4290), or uuid.Nil. A leg settles its
	// run (settleRunLegLocked) rather than running the rest of the card
	// itself: the run's own continuation does that, once, after the last
	// leg.
	Run uuid.UUID
}

// SetRuleFunc is a registered set rule: is `picked`, the live cards in
// the order submitted, a set the clause allows? It is never called for
// an empty pick and is a function of the cards alone, so it runs under
// the enumerator's read lock and must capture nothing.
type SetRuleFunc func(picked []Card) bool

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

// chooseCardsRules are the set rules of the continuations registered
// with one (RegisterChooseCardsThenRule), under the continuation's key.
// Guarded by effectRegistryMu.
var chooseCardsRules = map[string]SetRuleFunc{}

// RegisterChooseCardsThen registers a keyed choose_cards continuation.
// Call it once, from a package-level var. Panics on a bad or duplicate
// key, as DelayedBody does. The key is an on-disk identity: never
// renamed, never reused.
func RegisterChooseCardsThen(key string, fn ChooseCardsFunc) ChooseCardsThen {
	return RegisterChooseCardsThenRule(key, nil, fn)
}

// RegisterChooseCardsThenRule is RegisterChooseCardsThen for a pick with
// a set rule (#4264): every answer to the prompt, and every answer the
// enumerator offers, is checked against `rule` (nil allows every set the
// bounds allow). The rule shares the continuation's key, so it needs no
// ledger line of its own: a binary that does not know the key refuses
// the file, and one that knows it has the rule.
func RegisterChooseCardsThenRule(key string, rule SetRuleFunc, fn ChooseCardsFunc) ChooseCardsThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("choose-cards continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: choose-cards continuation %q has no function", key))
	}
	chooseCardsThens[key] = fn
	if rule != nil {
		chooseCardsRules[key] = rule
	}
	return ChooseCardsThen{key: key}
}

// lookupChooseCardsRule is the set rule registered under key, nil when
// the continuation has none.
func lookupChooseCardsRule(key string) SetRuleFunc {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	return chooseCardsRules[key]
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

// isKeyedChooseCards reports whether c is a choose_cards pick, or a
// reveal pick (#4264, resolution_pick.go), with a registered continuation.
func isKeyedChooseCards(c *PendingChoice) bool {
	return c != nil && (c.Kind == PendingChoiceChooseCards || c.Kind == PendingChoiceRevealPick) &&
		c.PickThen != "" && c.chooseCardsResume == nil
}

// runKeyedChooseCardsLocked runs c's continuation with `picked`
// (answered false: nobody chose). Caller must hold g.mu (write).
func (g *Game) runKeyedChooseCardsLocked(c *PendingChoice, picked []uuid.UUID, answered bool) error {
	fn, ok := lookupChooseCardsThen(c.PickThen)
	if !ok {
		return fmt.Errorf("%w: choose-cards continuation %q", ErrUnknownEffectKey, c.PickThen)
	}
	// #4281: a pick a keyed branch queued runs under that branch's cause.
	return g.withCarriedCauseLocked(carriedCauseOfChoice(c), func() error {
		return fn(g, CardsChosen{
			Chooser:    c.Chooser,
			Source:     c.Source,
			From:       c.FromPlayer,
			Picked:     copyUUIDs(picked),
			Answered:   answered,
			Candidates: copyUUIDs(c.ChooseCards),
			Min:        c.ChooseMin,
			Max:        c.ChooseMax,
			Carry:      copyUUIDs(c.OptionCarry),
			Question:   c.Reason,
			Word:       c.PickMatchArg,
			Key:        c.PickThen,
			Offer:      cloneAnyPlayerOffer(c.AnyPlayerOffer),
			Asked:      c.Asked.StackItem(),
			Run:        c.promptRun,
		})
	})
}
