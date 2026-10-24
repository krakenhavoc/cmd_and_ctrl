package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// pay_unless_keyed.go — an "unless that player pays <cost>" prompt
// whose decline is a KEY rather than a closure (#4256), so a table
// waiting on the payer is a restore point.
//
// QueuePayUnlessForEffect parks the decline on the prompt's
// payUnlessFrame as a closure, with the parsed cost beside it, and a
// prompt with a frame is not written to a restore point. A table waiting
// on Rhystic Study's, Smothering Tithe's or Esper Sentinel's payer could
// be undone but not saved and restored.
//
// The keyed form carries no frame. What the answer needs is plain data
// the prompt already has or that rides on it: the payer (Chooser), the
// card that asked (Source), the cost as printed (PayCost, re-parsed when
// the answer comes in), the decline's key (PickThen) and the IDs the
// card put there (OptionCarry). payUnlessFrameOf builds the frame the
// resolver, the automatic answer and the departure table read from that
// data, so every path that answers a pay-unless answers a keyed one the
// same way. The key shares the effect-key namespace and the ledger
// ("payunless <key>"). A binary from before this change refuses a
// restore point naming a key it does not have (checkEffectKeys).
//
// Only the plain mana shape is keyed: a cost, one decline and nothing
// on the "pay". The non-mana payments, the upkeep anchor and the
// any-player offers keep their frames.

// PayUnlessAnswer is what a keyed pay-unless decline is handed.
type PayUnlessAnswer struct {
	// Chooser is the player who was asked to pay; Source the card that
	// asked.
	Chooser, Source uuid.UUID
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
}

// PayUnlessFunc is a registered decline. It runs with g.mu held, may
// queue prompts of its own, and must capture nothing.
type PayUnlessFunc func(g *Game, r PayUnlessAnswer) error

// PayUnlessElse names a registered decline. The key is unexported, so
// the only way to hold one is RegisterPayUnlessElse.
type PayUnlessElse struct{ key string }

// Key is the decline's on-disk key.
func (e PayUnlessElse) Key() string { return e.key }

// payUnlessElses is guarded by effectRegistryMu, with the other effect
// keys, because the keys share one namespace and one ledger.
var payUnlessElses = map[string]PayUnlessFunc{}

// RegisterPayUnlessElse registers a keyed pay-unless decline. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity: never renamed, never reused.
func RegisterPayUnlessElse(key string, fn PayUnlessFunc) PayUnlessElse {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("pay-unless decline", key)
	if fn == nil {
		panic(fmt.Sprintf("game: pay-unless decline %q has no function", key))
	}
	payUnlessElses[key] = fn
	return PayUnlessElse{key: key}
}

func lookupPayUnlessElse(key string) (PayUnlessFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := payUnlessElses[key]
	return fn, ok
}

// KnownPayUnlessElse reports whether this binary has a decline
// registered under key.
func KnownPayUnlessElse(key string) bool { _, ok := lookupPayUnlessElse(key); return ok }

// registeredPayUnlessKeysLocked lists the ledger lines, "payunless
// <key>". Caller holds effectRegistryMu.
func registeredPayUnlessKeysLocked() []string {
	out := make([]string, 0, len(payUnlessElses))
	for k := range payUnlessElses {
		out = append(out, "payunless "+k)
	}
	sort.Strings(out)
	return out
}

// keyedPayUnlessDecline is the decline as the frame's closure, built
// fresh from the prompt's data each time it is read.
func keyedPayUnlessDecline(key string, r PayUnlessAnswer) func(g *Game) error {
	return func(g *Game) error {
		fn, ok := lookupPayUnlessElse(key)
		if !ok {
			return fmt.Errorf("%w: pay-unless decline %q", ErrUnknownEffectKey, key)
		}
		return fn(g, r)
	}
}

// QueuePayUnlessKeyedForEffect is QueuePayUnlessForEffect with the
// decline as a registered key and `carry` the plain data it is handed:
// "you may draw a card unless that player pays {1}" (Rhystic Study).
// The prompt carries no closure, so a table waiting on it is a restore
// point. Everything else is QueuePayUnlessForEffect's: an unparseable
// cost or a payer who has left runs the decline at once.
//
// Caller must hold g.mu.
func (g *Game) QueuePayUnlessKeyedForEffect(
	chooser, source uuid.UUID,
	cost, question string,
	onDecline PayUnlessElse,
	carry []uuid.UUID,
) error {
	if onDecline.key == "" {
		return fmt.Errorf("%w: keyed pay-unless with no decline", ErrInvalidParam)
	}
	decline := keyedPayUnlessDecline(onDecline.key, PayUnlessAnswer{Chooser: chooser, Source: source, Carry: copyUUIDs(carry)})
	if _, err := ParseCost(cost); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   source,
			ErrorMsg: "pay-unless: unparseable cost " + cost + ": " + err.Error(),
		})
		return decline(g)
	}
	if p := g.playerByIDLocked(chooser); p == nil || p.Eliminated {
		return decline(g)
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoicePayUnless,
		Chooser:     chooser,
		Count:       1,
		Source:      source,
		Reason:      question,
		PayCost:     cost,
		PickThen:    onDecline.key,
		OptionCarry: copyUUIDs(carry),
	})
	return nil
}

// payUnlessFrameOf is the frame a pay-unless prompt answers through:
// its own, or, for a keyed one, a frame built from its data (the cost
// re-parsed, the decline looked up by its key when it runs). Nil for any
// other prompt. Every reader of a pay-unless frame goes through this, so
// a keyed prompt is answered, auto-answered and dropped as a framed one
// is.
func payUnlessFrameOf(c *PendingChoice) *payUnlessFrame {
	if c == nil {
		return nil
	}
	if c.payUnlessResume != nil {
		return c.payUnlessResume
	}
	if c.Kind != PendingChoicePayUnless || c.PickThen == "" {
		return nil
	}
	parsed, err := ParseCost(c.PayCost)
	if err != nil {
		return nil
	}
	return &payUnlessFrame{
		cost: parsed,
		onDecline: keyedPayUnlessDecline(c.PickThen, PayUnlessAnswer{
			Chooser: c.Chooser,
			Source:  c.Source,
			Carry:   copyUUIDs(c.OptionCarry),
		}),
	}
}
