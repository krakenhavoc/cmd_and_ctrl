package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// may_cast_keyed.go — a may_cast prompt ("you may cast <card> without
// paying its mana cost", "you may play that land") whose answer is a KEY
// rather than a pair of closures (#4361), so a table waiting on it is a
// restore point.
//
// The closure form (QueueMayCastForEffect, and MayCastPrompt's OnAccept /
// OnDecline) held the two branches in a mayCastFrame on the prompt, and a
// prompt with a frame is not written to a restore point: a table waiting
// on hideaway's "play that land?" could be undone but not saved and
// restored. It is gone; this is the only door.
//
// The answer is a package-level function registered under an on-disk key
// (RegisterMayCastThen), handed the answer (Accepted) and what it needs as
// plain data the prompt already has: the chooser, the card that asked
// (Source), the card on offer (MayCastCard) and the IDs the queuing card
// put there (OptionCarry). On the prompt the key is PendingChoice.PickThen,
// the field every keyed prompt shares. The key shares the effect-key
// namespace and the ledger ("maycast <key>"). A binary from before this
// change refuses a restore point naming a key it does not have
// (checkEffectKeys).

// MayCastAnswer is what a keyed may_cast's answer is handed.
type MayCastAnswer struct {
	// Chooser answered; Source is the card that asked; Card is the card
	// the prompt offered (PendingChoice.MayCastCard).
	Chooser, Source, Card uuid.UUID
	// Accepted is the answer: true for the accept branch. A chooser who is
	// no longer in the game is answered false, at once, without a prompt.
	Accepted bool
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
}

// MayCastFunc is a registered answer. It runs with g.mu held, may queue
// prompts of its own, and must capture nothing.
type MayCastFunc func(g *Game, r MayCastAnswer) error

// MayCastThen names a registered may_cast answer. The key is unexported,
// so the only way to hold one is RegisterMayCastThen.
type MayCastThen struct{ key string }

// Key is the answer's on-disk key.
func (m MayCastThen) Key() string { return m.key }

// mayCastThens is guarded by effectRegistryMu, with the other effect keys,
// because the keys share one namespace and one ledger.
var mayCastThens = map[string]MayCastFunc{}

// RegisterMayCastThen registers a keyed may_cast answer. Call it once,
// from a package-level var. Panics on a bad or duplicate key. The key is
// an on-disk identity: never renamed, never reused.
func RegisterMayCastThen(key string, fn MayCastFunc) MayCastThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("may-cast answer", key)
	if fn == nil {
		panic(fmt.Sprintf("game: may-cast answer %q has no function", key))
	}
	mayCastThens[key] = fn
	return MayCastThen{key: key}
}

func lookupMayCastThen(key string) (MayCastFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := mayCastThens[key]
	return fn, ok
}

// KnownMayCastThen reports whether this binary has an answer registered
// under key.
func KnownMayCastThen(key string) bool { _, ok := lookupMayCastThen(key); return ok }

// registeredMayCastKeysLocked lists the ledger lines, "maycast <key>".
// Caller holds effectRegistryMu.
func registeredMayCastKeysLocked() []string {
	out := make([]string, 0, len(mayCastThens))
	for k := range mayCastThens {
		out = append(out, "maycast "+k)
	}
	sort.Strings(out)
	return out
}

// keyedMayCastBranch runs one branch of a keyed may_cast: the answer
// looked up by its key, under the cause the prompt carried.
func (g *Game) keyedMayCastBranch(key string, r MayCastAnswer, cause carriedCause) error {
	fn, ok := lookupMayCastThen(key)
	if !ok {
		return fmt.Errorf("%w: may-cast answer %q", ErrUnknownEffectKey, key)
	}
	return g.withCarriedCauseLocked(cause, func() error { return fn(g, r) })
}

// QueueMayCastKeyedForEffect queues a may_cast prompt for p.Chooser, asked
// from inside a resolving ability: the answer is a registered key and
// `carry` the plain data it is handed. The prompt carries no closure, so a
// table waiting on it is a restore point.
//
// A chooser who is no longer seated / is eliminated is not asked and the
// decline runs instead — the same default pay-unless takes, and right for
// the same reason: the consequence attached to the answer that can no
// longer be given is the one that happens.
//
// Caller must hold g.mu.
func (g *Game) QueueMayCastKeyedForEffect(p MayCastPrompt, then MayCastThen, carry []uuid.UUID) error {
	if then.key == "" {
		return fmt.Errorf("%w: keyed may-cast with no answer", ErrInvalidParam)
	}
	if pl := g.playerByIDLocked(p.Chooser); pl == nil || pl.Eliminated {
		return g.keyedMayCastBranch(then.key, MayCastAnswer{
			Chooser: p.Chooser, Source: p.Source, Card: p.Card, Carry: copyUUIDs(carry),
		}, carriedCause{})
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:           PendingChoiceMayCast,
		Chooser:        p.Chooser,
		Count:          1,
		Source:         p.Source,
		Reason:         p.Question,
		MayCastCard:    p.Card,
		MayCastKeyword: p.Keyword,
		AcceptLabel:    p.AcceptLabel,
		DeclineLabel:   p.DeclineLabel,
		PickThen:       then.key,
		OptionCarry:    copyUUIDs(carry),
		private:        p.Private,
	})
	return nil
}

// mayCastBranchOf is the branch a may_cast answer runs, built from the
// prompt's data (the answer looked up by its key when it runs). Nil for a
// prompt with no key.
func mayCastBranchOf(c *PendingChoice, accepted bool) func(g *Game) error {
	if c == nil || c.Kind != PendingChoiceMayCast || c.PickThen == "" {
		return nil
	}
	key := c.PickThen
	r := MayCastAnswer{
		Chooser: c.Chooser, Source: c.Source, Card: c.MayCastCard,
		Accepted: accepted, Carry: copyUUIDs(c.OptionCarry),
	}
	// #4281: a prompt a keyed branch queued answers under that branch's
	// cause.
	cause := carriedCauseOfChoice(c)
	return func(g *Game) error { return g.keyedMayCastBranch(key, r, cause) }
}
