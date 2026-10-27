package game

import (
	"fmt"

	"github.com/google/uuid"
)

// outside_pick_keyed.go — an outside-the-game pick whose filter and
// continuation are KEYS rather than closures (#4256), so a table waiting
// on one is a restore point.
//
// QueueOutsidePickForEffect has two branches (ADR 0191, ADR 0198): a
// player whose deck seated a sideboard chooses from it with a
// choose_cards prompt, and a player whose deck seated none chooses from
// the card database with a PendingChoiceOutsideCollection, answered with
// card names. Both resumed through a closure: the choose_cards frame's
// Then, and the collection pick's searchResumeFrame, which also held the
// filter the typed names are checked against. A table waiting on either
// could be undone but not saved and restored.
//
// Since #4361 S4 every outside pick is keyed: the closure form (Match,
// Validate, Then) had no production caller and is gone, and a pick that
// names no key of its own (Wish's "name a card to play", which only
// mints the card into the container) rides outsidePickPlain.
//
// A keyed pick registers both halves under one key
// (RegisterOutsidePick): the continuation, as a keyed choose-cards
// continuation (choose_cards_keyed.go, ledgered "choose <key>"), and the
// filter, as a function of the card and one plain word (MatchArg, a
// text-changed subtype for Spawnsire of Ulamog). On the sideboard branch
// the filter is read once, when the candidates are listed, and the
// prompt is an ordinary keyed choose_cards pick. On the database branch
// the prompt carries the key (PendingChoice.PickThen), the word
// (PickMatchArg) and the carried IDs (OptionCarry); the answer looks the
// filter up by the key, checks the typed names against it, mints the
// cards and runs the continuation with their new IDs, exactly as the
// closure form does.
//
// A set rule (OutsidePickPrompt.Validate, Extrapolate the Impossible's
// "two cards with different names") is registered with the pick
// (RegisterOutsidePickRule, #4264), under the same key, and looked up by
// it on both branches. A continuation re-checks where its picks are now
// (StillOutsideForEffect): a keyed pick has no live-zone re-check of its
// own.

// OutsideMatchFunc is a registered outside-pick filter: does `c` satisfy
// the clause, given the word the prompt carries? Nil offers every card.
// It runs under the game lock, on the database branch inside the
// enumerator's read lock too, so it is a function of the card and the
// word alone and must capture nothing.
type OutsideMatchFunc func(c Card, arg string) bool

// OutsidePick names a registered outside-the-game pick. The key is
// unexported, so the only way to hold one is RegisterOutsidePick.
type OutsidePick struct{ then ChooseCardsThen }

// Key is the pick's on-disk key.
func (p OutsidePick) Key() string { return p.then.key }

// outsidePickEntry is one registered pick's filter. A struct so a nil
// filter ("every card") is still a registered entry.
type outsidePickEntry struct{ match OutsideMatchFunc }

// outsidePicks is guarded by effectRegistryMu. The keys are the
// continuations' keys in chooseCardsThens, so they share its namespace
// and its ledger lines.
var outsidePicks = map[string]outsidePickEntry{}

// RegisterOutsidePick registers a keyed outside-the-game pick: its
// filter (nil for every card) and its continuation, which is handed the
// picked cards' IDs in the chooser's outside-the-game container. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity: never renamed, never reused, ledgered as
// "choose <key>".
func RegisterOutsidePick(key string, match OutsideMatchFunc, then ChooseCardsFunc) OutsidePick {
	return RegisterOutsidePickRule(key, match, nil, then)
}

// RegisterOutsidePickRule is RegisterOutsidePick for a pick with a set
// rule over the whole answer (#4264): "exactly two cards with different
// names". The rule is never called for an empty pick.
func RegisterOutsidePickRule(key string, match OutsideMatchFunc, rule SetRuleFunc, then ChooseCardsFunc) OutsidePick {
	t := RegisterChooseCardsThenRule(key, rule, then)
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	outsidePicks[key] = outsidePickEntry{match: match}
	return OutsidePick{then: t}
}

// lookupOutsideMatch is the filter registered under key, and whether the
// key names an outside pick.
func lookupOutsideMatch(key string) (OutsideMatchFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	e, ok := outsidePicks[key]
	return e.match, ok
}

// KnownOutsidePick reports whether this binary has an outside pick, its
// filter and its continuation, registered under key.
func KnownOutsidePick(key string) bool {
	_, ok := lookupOutsideMatch(key)
	return ok && KnownChooseCardsThen(key)
}

// boundOutsideMatch is a registered filter with its word bound, in the
// shape the closure form's Match has. Nil stays nil.
func boundOutsideMatch(fn OutsideMatchFunc, arg string) func(Card) bool {
	if fn == nil {
		return nil
	}
	return func(c Card) bool { return fn(c, arg) }
}

// outsidePickPlain is the pick of an outside-the-game prompt that names
// no key of its own: every card, no set rule, and nothing after the pick
// (the answer itself is the effect: a collection pick mints the named
// card into the container). An on-disk key: never renamed, never reused.
var outsidePickPlain = RegisterOutsidePick("engine/outside-pick-plain", nil,
	func(*Game, CardsChosen) error { return nil })

// outsidePickMatch is the filter a prompt queues with: the keyed pick's
// registered one, its word bound. p.Keyed must be set.
func outsidePickMatch(p OutsidePickPrompt) (func(Card) bool, error) {
	key := p.Keyed.Key()
	fn, ok := lookupOutsideMatch(key)
	if !ok {
		return nil, fmt.Errorf("%w: outside pick %q", ErrUnknownEffectKey, key)
	}
	return boundOutsideMatch(fn, p.MatchArg), nil
}

// collectionFilterLocked is a collection pick's per-card filter and set
// rule: the pick's registered filter and set rule. ok is false for a
// choice that names no registered pick.
func collectionFilterLocked(choice *PendingChoice) (pred func(Card) bool, validate func([]Card) bool, ok bool) {
	if choice == nil || choice.PickThen == "" {
		return nil, nil, false
	}
	fn, known := lookupOutsideMatch(choice.PickThen)
	if !known {
		return nil, nil, false
	}
	return boundOutsideMatch(fn, choice.PickMatchArg), lookupChooseCardsRule(choice.PickThen), true
}

// StillOutsideForEffect is the cards of `ids` still in `player`'s
// outside-the-game container, in the order given: a keyed pick's
// continuation's re-check, since the prompt has none of its own.
//
// Caller must hold g.mu.
func (g *Game) StillOutsideForEffect(player uuid.UUID, ids []uuid.UUID) []uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Sideboard == nil {
		return nil
	}
	var out []uuid.UUID
	for _, id := range ids {
		if p.Sideboard.Contains(id) {
			out = append(out, id)
		}
	}
	return out
}
