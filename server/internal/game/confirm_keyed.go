package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// confirm_keyed.go — a two-way confirm whose answer is a KEY rather than
// a pair of closures (#4264), so a table waiting on it is a restore
// point.
//
// The closure form (QueueConfirmForEffect, deleted in #4253's last part)
// held the two branches in a confirmFrame, and a prompt with a frame is not
// written to a restore point: a table waiting on Rhystic Study's "draw a
// card?" could be undone but not saved and restored.
//
// The keyed form has no frame. The answer is a package-level function
// registered under an on-disk key (RegisterConfirmThen), handed the
// answer (Accepted) and what it needs as plain data the prompt already
// has or that rides on it: the chooser, the card that asked (Source) and
// the IDs the card put there (OptionCarry). On the prompt the key is
// PendingChoice.PickThen, the field every keyed prompt shares. The key
// shares the effect-key namespace and the ledger ("confirm <key>"). A
// binary from before this change refuses a restore point naming a key it
// does not have (checkEffectKeys).
//
// confirmFrameOf builds, from that data, the frame the resolver and the
// automatic answer read, so every path that answers a confirm answers a
// keyed one the same way.

// ConfirmAnswer is what a keyed confirm's answer is handed.
type ConfirmAnswer struct {
	// Chooser answered; Source is the card that asked.
	Chooser, Source uuid.UUID
	// Accepted is the answer: true for the accept branch.
	Accepted bool
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
	// LifeCost is the life the accept branch was declared to charge
	// (ConfirmPrompt.LifeCost): "pay N life" is a number a shared answer
	// reads off the prompt rather than from a key of its own.
	LifeCost int
	// PoisonCost is the poison counters the accept branch was declared to give
	// the chooser (ConfirmPrompt.PoisonCost, #4613).
	PoisonCost int
	// Word is the one word of the card's text the queuing card put on the
	// prompt (ConfirmPrompt.Word, #4291): an amount or a token's name that a
	// shared answer ("draw N cards", "create a token") is parameterised by.
	// Empty when none.
	Word string
	// Objects are the objects, with the epoch each had, the card named as it
	// asked (ConfirmPrompt.Objects), so an answer can tell a permanent from a
	// new object the same card has become (CR 400.7).
	Objects []ObjectRef
	// Item is the stack item that asked, rebuilt from the data the prompt
	// carried (ConfirmPrompt.Asked), so an answer reads the trigger's event,
	// the targets, the X and the source object's epoch as a closure over the
	// item used to. Nil when the card put none there.
	Item *StackItem
}

// ConfirmFunc is a registered answer. It runs with g.mu held, may queue
// prompts of its own, and must capture nothing.
type ConfirmFunc func(g *Game, r ConfirmAnswer) error

// ConfirmThen names a registered confirm answer. The key is unexported,
// so the only way to hold one is RegisterConfirmThen.
type ConfirmThen struct{ key string }

// Key is the answer's on-disk key.
func (c ConfirmThen) Key() string { return c.key }

// confirmThens is guarded by effectRegistryMu, with the other effect
// keys, because the keys share one namespace and one ledger.
var confirmThens = map[string]ConfirmFunc{}

// RegisterConfirmThen registers a keyed confirm answer. Call it once,
// from a package-level var. Panics on a bad or duplicate key. The key is
// an on-disk identity: never renamed, never reused.
func RegisterConfirmThen(key string, fn ConfirmFunc) ConfirmThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("confirm answer", key)
	if fn == nil {
		panic(fmt.Sprintf("game: confirm answer %q has no function", key))
	}
	confirmThens[key] = fn
	return ConfirmThen{key: key}
}

func lookupConfirmThen(key string) (ConfirmFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := confirmThens[key]
	return fn, ok
}

// KnownConfirmThen reports whether this binary has an answer registered
// under key.
func KnownConfirmThen(key string) bool { _, ok := lookupConfirmThen(key); return ok }

// registeredConfirmKeysLocked lists the ledger lines, "confirm <key>".
// Caller holds effectRegistryMu.
func registeredConfirmKeysLocked() []string {
	out := make([]string, 0, len(confirmThens))
	for k := range confirmThens {
		out = append(out, "confirm "+k)
	}
	sort.Strings(out)
	return out
}

// keyedConfirmBranch is one branch of a keyed confirm as the frame's
// closure, built fresh from the prompt's data each time it is read.
func keyedConfirmBranch(key string, r ConfirmAnswer, cause carriedCause) func(g *Game) error {
	return func(g *Game) error {
		fn, ok := lookupConfirmThen(key)
		if !ok {
			return fmt.Errorf("%w: confirm answer %q", ErrUnknownEffectKey, key)
		}
		return g.withCarriedCauseLocked(cause, func() error { return fn(g, r) })
	}
}

// RunConfirmThenForEffect runs a keyed answer directly, for the caller
// that has decided the question need not be asked and still has the
// rest of the card to run (a "you may draw" that a draw limit forbids
// runs the decline at once). Caller must hold g.mu.
func (g *Game) RunConfirmThenForEffect(then ConfirmThen, r ConfirmAnswer) error {
	if then.key == "" {
		return nil
	}
	return keyedConfirmBranch(then.key, r, carriedCause{})(g)
}

// QueueConfirmKeyedForEffect queues a two-way prompt and returns its ID: the
// answer is a registered key and `carry` the plain data it is handed. The
// prompt carries no closure, so a table waiting on it is a restore point.
//
// Caller must hold g.mu.
func (g *Game) QueueConfirmKeyedForEffect(p ConfirmPrompt, then ConfirmThen, carry []uuid.UUID) (uuid.UUID, error) {
	if then.key == "" {
		return uuid.Nil, fmt.Errorf("%w: keyed confirm with no answer", ErrInvalidParam)
	}
	from := p.FromPlayer
	if from == uuid.Nil {
		from = p.Chooser
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceConfirm,
		Chooser:      p.Chooser,
		FromPlayer:   from,
		Count:        1,
		Source:       p.Source,
		Reason:       p.Question,
		AcceptLabel:  p.AcceptLabel,
		DeclineLabel: p.DeclineLabel,
		LifeCost:     p.LifeCost,
		PoisonCost:   p.PoisonCost,

		DeclineEndsLoop: p.DeclineEndsLoop,
		PickThen:        then.key,
		OptionCarry:     copyUUIDs(carry),
		PickMatchArg:    p.Word,
		OptionObjects:   cloneObjectRefs(p.Objects),
		Asked:           copyAskedItem(p.Asked),
	}), nil
}

// confirmFrameOf is the frame a confirm answers through: its own, or,
// for a keyed one, a frame built from its data (the answer looked up by
// its key when it runs). Nil for any other prompt. Every reader of a
// confirm frame goes through this.
func confirmFrameOf(c *PendingChoice) *confirmFrame {
	if c == nil {
		return nil
	}
	if c.Kind != PendingChoiceConfirm || c.PickThen == "" {
		return nil
	}
	r := ConfirmAnswer{
		Chooser: c.Chooser, Source: c.Source, Carry: copyUUIDs(c.OptionCarry),
		LifeCost: c.LifeCost, PoisonCost: c.PoisonCost, Word: c.PickMatchArg, Objects: cloneObjectRefs(c.OptionObjects),
		Item: c.Asked.StackItem(),
	}
	accepted := r
	accepted.Accepted = true
	// #4281: a confirm a keyed branch queued answers under that branch's
	// cause.
	cause := carriedCauseOfChoice(c)
	return &confirmFrame{
		onAccept:  keyedConfirmBranch(c.PickThen, accepted, cause),
		onDecline: keyedConfirmBranch(c.PickThen, r, cause),
	}
}
