package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// pay_unless_keyed.go — an "unless that player pays <cost>" prompt
// whose branches are a KEY rather than closures (#4256, #4264), so a
// table waiting on the payer is a restore point.
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
// the answer comes in), the branches' key (PickThen), the IDs the card
// put there (OptionCarry), and, for the shapes #4264 added, the step the
// prompt is owed in (OwedInStep), the stack object it guards
// (GuardsStackItem), the life paid beside the cost (LifeCost) and a
// non-mana payment (PayActionClause, plain data like the rest of
// PayAction). payUnlessFrameOf builds the frame the resolver, the
// automatic answer and the departure table read from that data, so every
// path that answers a pay-unless answers a keyed one the same way.
//
// One key names the card's two branches: the decline ("unless that
// player pays"), the payment ("you may pay. If you do") or both (echo's
// event, Springheart Nantuko). It shares the effect-key namespace and
// the ledger ("payunless <key>"). A binary from before this change
// refuses a restore point naming a key it does not have
// (checkEffectKeys).
//
// QueuePayUnlessPromptKeyedForEffect is the one door for every shape the
// keyed form covers: the plain pay-unless (Rhystic Study), the upkeep
// "pay or else" that holds its step (Stasis, cumulative upkeep), the
// "you may pay" (Hellkite Charger), a counter-unless-pays guarding a
// stack object (QueueCounterUnlessPaidForEffect's own default is keyed)
// and the non-mana payments (a discard, a sacrifice, a tap). The
// waterbend tap clause stays a closure.

// PayUnlessAnswer is what a keyed pay-unless branch is handed.
type PayUnlessAnswer struct {
	// Chooser is the player who was asked to pay; Source the card that
	// asked.
	Chooser, Source uuid.UUID
	// Carry is the plain data the queuing card put on the prompt.
	Carry []uuid.UUID
	// Guards is the stack object the prompt guarded, uuid.Nil for a
	// prompt that guards none (a counter-unless-pays names the spell it
	// counts on).
	Guards uuid.UUID
	// Item is the stack item that asked, rebuilt from the data the
	// prompt carried (AskedItem), or nil for a prompt whose card put none
	// there. A branch runs against it, so it reads the asking trigger's
	// event, targets and X, and the object (CR 400.7) the ability came
	// from, as a closure over the item used to.
	Item *StackItem
	// Objects are the objects, with their epochs, the card put on the
	// prompt beside Carry (PendingChoice.OptionObjects).
	Objects []ObjectRef
}

// PayUnlessFunc is a registered branch. It runs with g.mu held, may
// queue prompts of its own, and must capture nothing.
type PayUnlessFunc func(g *Game, r PayUnlessAnswer) error

// PayUnlessElse names a card's registered branches. The key is
// unexported, so the only way to hold one is RegisterPayUnlessElse or
// RegisterPayUnlessBranches.
type PayUnlessElse struct{ key string }

// Key is the branches' on-disk key.
func (e PayUnlessElse) Key() string { return e.key }

// payUnlessEntry is one registered key's branches: what runs when the
// payment is not made (decline) and when it is (pay). Either may be nil.
type payUnlessEntry struct{ decline, pay PayUnlessFunc }

// payUnlessElses is guarded by effectRegistryMu, with the other effect
// keys, because the keys share one namespace and one ledger.
var payUnlessElses = map[string]payUnlessEntry{}

// RegisterPayUnlessElse registers a keyed pay-unless decline. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity: never renamed, never reused.
func RegisterPayUnlessElse(key string, fn PayUnlessFunc) PayUnlessElse {
	if fn == nil {
		panic(fmt.Sprintf("game: pay-unless decline %q has no function", key))
	}
	return RegisterPayUnlessBranches(key, nil, fn)
}

// RegisterPayUnlessBranches registers both branches of a keyed
// pay-unless under one key (#4264): `onPay` runs once the payment has
// been made ("you may pay. If you do", echo's event) and `onDecline`
// when it has not. Either may be nil, but not both. Call it once, from a
// package-level var. Panics on a bad or duplicate key.
func RegisterPayUnlessBranches(key string, onPay, onDecline PayUnlessFunc) PayUnlessElse {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("pay-unless branches", key)
	if onPay == nil && onDecline == nil {
		panic(fmt.Sprintf("game: pay-unless branches %q have no function", key))
	}
	payUnlessElses[key] = payUnlessEntry{decline: onDecline, pay: onPay}
	return PayUnlessElse{key: key}
}

func lookupPayUnlessElse(key string) (payUnlessEntry, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	e, ok := payUnlessElses[key]
	return e, ok
}

// KnownPayUnlessElse reports whether this binary has branches
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

// keyedPayUnlessBranch is one branch of a keyed pay-unless as the
// frame's closure, looked up by its key each time it runs. A key this
// binary lacks is an error (checkEffectKeys refuses such a restore
// point, so this is the live-table guard); a branch the key does not
// have does nothing.
func keyedPayUnlessBranch(key string, r PayUnlessAnswer, pay bool, cause carriedCause) func(g *Game) error {
	return func(g *Game) error {
		e, ok := lookupPayUnlessElse(key)
		if !ok {
			return fmt.Errorf("%w: pay-unless branches %q", ErrUnknownEffectKey, key)
		}
		fn := e.decline
		if pay {
			fn = e.pay
		}
		if fn == nil {
			return nil
		}
		return g.withCarriedCauseLocked(cause, func() error { return fn(g, r) })
	}
}

// PayUnlessKeyedPrompt is the whole of a keyed pay-unless: the shapes
// differ only in these fields, and the prompt, the payment and the
// answer are pay-unless's own.
type PayUnlessKeyedPrompt struct {
	// Chooser is the player asked; Source the card that asks.
	Chooser, Source uuid.UUID
	// Cost is the printed payment, parsed by ParseCost. With an Action
	// it is the payment's words ("Discard a card") and never parsed.
	Cost string
	// Question is the dialog header.
	Question string
	// Then names the card's branches (RegisterPayUnlessBranches). The
	// branches are handed Carry.
	Then  PayUnlessElse
	Carry []uuid.UUID
	// Action is a payment that is not mana (ADR 0108 §5): plain data on
	// the prompt (PendingChoice.PayActionClause). Nil for mana.
	Action *PayAction
	// InThisStep holds the table in the step the prompt was queued in
	// until it is answered (OwedInStep, #997): the upkeep "pay or else",
	// and a "you may pay" about the phase in progress.
	InThisStep bool
	// MayPay is "you may pay. If you do, …": a payer who cannot be asked,
	// or a cost that cannot be read, runs nothing, where a pay-unless
	// runs its decline (the answer that can no longer be given is the
	// payment).
	MayPay bool
	// Guards is the stack object the answer decides, held on the stack
	// while the prompt is open (GuardsStackItem, #951), or uuid.Nil.
	Guards uuid.UUID
	// LifeCost is mandatory life paid with the cost in one payment.
	LifeCost int
	// Asked is the asking stack item as data (NewAskedItem), handed to the
	// branches as PayUnlessAnswer.Item. Nil when they read none.
	Asked *AskedItem
	// Objects are objects with their epochs the branches are handed
	// (PayUnlessAnswer.Objects), for a permanent the card named as the
	// prompt was queued.
	Objects []ObjectRef
}

// copyPayAction is a deep copy of a non-mana payment, nil for nil.
func copyPayAction(a *PayAction) *PayAction {
	if a == nil {
		return nil
	}
	out := *a
	if len(a.Either) > 0 {
		out.Either = append([]PayAction(nil), a.Either...)
	}
	return &out
}

// QueuePayUnlessPromptKeyedForEffect queues a keyed pay-unless of any
// shape. The prompt carries no closure, so a table waiting on it is a
// restore point. An unparseable cost or a payer who has left runs the
// decline at once (nothing, for a MayPay).
//
// Caller must hold g.mu.
func (g *Game) QueuePayUnlessPromptKeyedForEffect(p PayUnlessKeyedPrompt) error {
	if p.Then.key == "" {
		return fmt.Errorf("%w: keyed pay-unless with no branches", ErrInvalidParam)
	}
	if p.LifeCost < 0 {
		return ErrInvalidParam
	}
	r := PayUnlessAnswer{Chooser: p.Chooser, Source: p.Source, Carry: copyUUIDs(p.Carry), Guards: p.Guards, Item: p.Asked.StackItem(), Objects: cloneObjectRefs(p.Objects)}
	// #4279: the cause is taken while the asking resolution is open.
	cause := carriedCause{MoveCause: g.resolutionCauseLocked(), Source: p.Source}
	decline := func(g *Game) error {
		if p.MayPay {
			return nil
		}
		return keyedPayUnlessBranch(p.Then.key, r, false, cause)(g)
	}
	var err error
	switch {
	case p.Action == nil:
		_, err = ParseCost(p.Cost)
	case p.Action.Kind == PayActionEither && !validEither(p.Action):
		// ADR 0187 §5: an either payment needs two or more branches,
		// none of them an either. A programming error, surfaced once;
		// the consequence still runs so the card does something.
		err = ErrInvalidParam
	}
	if err != nil {
		what := "pay-unless"
		if p.MayPay {
			what = "may-pay"
		}
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   p.Source,
			ErrorMsg: what + ": unparseable cost " + p.Cost + ": " + err.Error(),
		})
		return decline(g)
	}
	if pl := g.playerByIDLocked(p.Chooser); pl == nil || pl.Eliminated {
		return decline(g)
	}
	var owed TurnStep
	if p.InThisStep {
		owed = g.currentTurnStepLocked()
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:            PendingChoicePayUnless,
		Chooser:         p.Chooser,
		Count:           1,
		Source:          p.Source,
		Reason:          p.Question,
		PayCost:         p.Cost,
		LifeCost:        p.LifeCost,
		OwedInStep:      owed,
		GuardsStackItem: p.Guards,
		PayActionClause: copyPayAction(p.Action),
		Asked:           copyAskedItem(p.Asked),
		OptionObjects:   cloneObjectRefs(p.Objects),
		PickThen:        p.Then.key,
		OptionCarry:     copyUUIDs(p.Carry),

		PayCauseController: cause.Controller,
		PayCauseItem:       cause.Item,
	})
	return nil
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
	return g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
		Chooser: chooser, Source: source, Cost: cost, Question: question,
		Then: onDecline, Carry: carry,
	})
}

// payUnlessFrameOf is the frame a pay-unless prompt answers through:
// its own, or, for a keyed one, a frame built from its data (the cost
// re-parsed or the non-mana payment read off the prompt, the branches
// looked up by their key when they run). Nil for any other prompt.
// Every reader of a pay-unless frame goes through this, so a keyed
// prompt is answered, auto-answered and dropped as a framed one is.
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
	if c.AnyPlayerOffer != nil {
		// #4264: a keyed any-player leg rebuilds the offer's next step.
		return anyPlayerFrameOf(c)
	}
	f := &payUnlessFrame{}
	if c.PayActionClause != nil {
		f.action = copyPayAction(c.PayActionClause)
	} else {
		parsed, err := ParseCost(c.PayCost)
		if err != nil {
			return nil
		}
		f.cost = parsed
	}
	r := PayUnlessAnswer{
		Chooser: c.Chooser,
		Source:  c.Source,
		Carry:   copyUUIDs(c.OptionCarry),
		Guards:  c.GuardsStackItem,
		Item:    c.Asked.StackItem(),
		Objects: cloneObjectRefs(c.OptionObjects),
	}
	cause := carriedCauseOfChoice(c)
	// A branch the key does not have stays nil, so the engine's own "this
	// branch does nothing" reading holds (a may-pay's decline, a plain
	// pay-unless's payment). A key this binary lacks keeps the decline,
	// which reports it.
	e, known := lookupPayUnlessElse(c.PickThen)
	if !known || e.decline != nil {
		f.onDecline = keyedPayUnlessBranch(c.PickThen, r, false, cause)
	}
	if known && e.pay != nil {
		f.onPay = keyedPayUnlessBranch(c.PickThen, r, true, cause)
	}
	return f
}

// QueueMayPayKeyedForEffect is QueueMayPayForEffect (and, with
// inThisStep, QueueMayPayInThisStepForEffect) with the payment's
// consequence as a registered key (#4264): "you may pay {2}{U}. If you
// do, ...". The key's pay branch runs only when the mana was really
// paid; a payer who cannot be asked runs nothing. The prompt carries no
// closure, so a table waiting on it is a restore point.
//
// Caller must hold g.mu.
func (g *Game) QueueMayPayKeyedForEffect(
	chooser, source uuid.UUID,
	cost, question string,
	then PayUnlessElse,
	carry []uuid.UUID,
	inThisStep bool,
) error {
	return g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
		Chooser: chooser, Source: source, Cost: cost, Question: question,
		Then: then, Carry: carry, InThisStep: inThisStep, MayPay: true,
	})
}

// QueueMayPayElseKeyedForEffect is QueueMayPayElseForEffect with both
// consequences as one registered key: "you may pay {1}. If you do, ....
// If you don't, ..." (Springheart Nantuko). A payer who cannot be asked
// runs the decline, pay-unless's default.
//
// Caller must hold g.mu.
func (g *Game) QueueMayPayElseKeyedForEffect(
	chooser, source uuid.UUID,
	cost, question string,
	then PayUnlessElse,
	carry []uuid.UUID,
) error {
	return g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
		Chooser: chooser, Source: source, Cost: cost, Question: question,
		Then: then, Carry: carry,
	})
}
