package game

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/google/uuid"
)

// number_then.go — a number-answer prompt whose consequence is a KEY
// rather than a closure (#4294), so a table waiting on the answer is a
// restore point.
//
// A pay_amount prompt (PendingChoicePayAmount) asks a player for a number:
// "you may pay any amount of {E}" (Harnessed Lightning), "an amount of
// damage of your choice" (Volcano Hellion), "how many of the tokens attack
// this player" (Tilonalli's Summoner), an assist helper's share of a
// spell's generic mana. Until now every one of them held the rest of the
// card as a Go closure on the prompt's chooseValueResume frame, and a
// prompt with a frame is not written to a restore point (#4289 keyed only
// the "choose X" half of a MayPayX, whose branches the payment that
// follows runs).
//
// The keyed form carries no frame. The branch is a package-level function
// registered under an on-disk key (RegisterNumberThen), handed the number
// and what it needs as plain data the prompt carries: the chooser and the
// card that asked (Chooser, Source), the IDs the card put there (Carry,
// PendingChoice.OptionCarry), a few words and numbers of its own (Words,
// Nums: PickWords, PickNums), and the stack item that asked, rebuilt from
// AskedItem. Share a key where the consequence is the same: "deal that
// much damage to the target" is one key for every card that says it.
//
// On the prompt the key is PendingChoice.PickThen, the field every keyed
// prompt shares, so a binary from before this change refuses a restore
// point naming a key it does not have (checkEffectKeys reads a pay_amount
// without a MayPayX clause as a number prompt). The key shares the
// effect-key namespace and the ledger ("number <key>").
//
// numberFrameOf builds, from the prompt's data, the frame the resolver and
// the departure rule read (chooseValueFrameOf), so every path that
// answers a pay_amount answers a keyed one the same way. The answer runs
// under the cause the prompt was queued under (#4279), as a keyed
// pay-unless's does: the table has moved on by the time a restored prompt
// is answered.

// NumberAnswer is what a keyed number branch is handed.
type NumberAnswer struct {
	// Chooser answered the prompt; Source is the card that asked.
	Chooser, Source uuid.UUID
	// Number is the answer: what was paid for a payment (0 when nothing
	// was, and when the chooser left), and for a number that is not paid
	// the answer, or its floor when the chooser left.
	Number int
	// Carry, Words and Nums are the plain data the queuing card put on
	// the prompt. A chain of questions (one per kind of counter) keeps
	// its place in them.
	Carry []uuid.UUID
	Words []string
	Nums  []int
	// Item is the stack item that asked, rebuilt from the data the prompt
	// carried (AskedItem), or nil for a prompt whose card put none there
	// (a permanent's "as it enters" choice). A branch runs against it, so
	// it reads the asking trigger's event, the targets and the X as a
	// closure over the item used to.
	Item *StackItem

	// assist and split are the engine's own branches' data (assist.go,
	// attack_split.go).
	assist *AssistPrompt
	split  *AttackSplit
}

// NumberFunc is a registered branch. It runs with g.mu held, may queue
// prompts of its own, and must capture nothing.
type NumberFunc func(g *Game, r NumberAnswer) error

// NumberThen names a registered number branch. The key is unexported, so
// the only way to hold one is RegisterNumberThen.
type NumberThen struct{ key string }

// Key is the branch's on-disk key.
func (n NumberThen) Key() string { return n.key }

// numberThens is guarded by effectRegistryMu, with the other effect keys,
// because the keys share one namespace and one ledger.
var numberThens = map[string]NumberFunc{}

// RegisterNumberThen registers a keyed number branch. Call it once, from a
// package-level var. Panics on a bad or duplicate key. The key is an
// on-disk identity: never renamed, never reused.
func RegisterNumberThen(key string, fn NumberFunc) NumberThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("number branch", key)
	if fn == nil {
		panic(fmt.Sprintf("game: number branch %q has no function", key))
	}
	numberThens[key] = fn
	return NumberThen{key: key}
}

func lookupNumberThen(key string) (NumberFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := numberThens[key]
	return fn, ok
}

// KnownNumberThen reports whether this binary has a branch registered
// under key.
func KnownNumberThen(key string) bool { _, ok := lookupNumberThen(key); return ok }

// registeredNumberKeysLocked lists the ledger lines, "number <key>".
// Caller holds effectRegistryMu.
func registeredNumberKeysLocked() []string {
	out := make([]string, 0, len(numberThens))
	for k := range numberThens {
		out = append(out, "number "+k)
	}
	sort.Strings(out)
	return out
}

// NumberBranch is the keyed consequence of a number prompt and the plain
// data it is handed: the part of ChooseNumber and PayEnergyAmount that
// replaces their closures. The zero value is "no branch".
type NumberBranch struct {
	// Then is the registered branch.
	Then NumberThen
	// Carry, Words and Nums are the data the branch is handed
	// (NumberAnswer). Asked is the asking stack item as data
	// (NewAskedItem).
	Carry []uuid.UUID
	Words []string
	Nums  []int
	Asked *AskedItem
}

// IsZero reports whether no branch is named.
func (b NumberBranch) IsZero() bool { return b.Then.key == "" }

// answer is the NumberAnswer for the number n.
func (b NumberBranch) answer(chooser, source uuid.UUID, n int) NumberAnswer {
	return NumberAnswer{
		Chooser: chooser, Source: source, Number: n,
		Carry: copyUUIDs(b.Carry), Words: copyStrings(b.Words), Nums: copyInts(b.Nums),
		Item: b.Asked.StackItem(),
	}
}

// runNumberBranch runs the branch the key names, under `cause`. A key this
// binary lacks is an error (checkEffectKeys refuses such a restore point,
// so this is the live-table guard).
//
// Caller must hold g.mu.
func (g *Game) runNumberBranch(key string, r NumberAnswer, cause carriedCause) error {
	fn, ok := lookupNumberThen(key)
	if !ok {
		return fmt.Errorf("%w: number branch %q", ErrUnknownEffectKey, key)
	}
	return g.withCarriedCauseLocked(cause, func() error { return fn(g, r) })
}

// runNumberBranchNow runs a branch for a prompt that was never asked
// (nothing to choose, or the chooser has left) with `n`, inside the
// resolution that would have asked it.
//
// Caller must hold g.mu.
func (g *Game) runNumberBranchNow(b NumberBranch, chooser, source uuid.UUID, n int) error {
	return g.runNumberBranch(b.Then.key, b.answer(chooser, source, n), carriedCause{})
}

// RunNumberThenForEffect runs a branch directly, for the caller that has
// decided the question need not be asked and still has the rest of the
// card to run. Caller must hold g.mu.
func (g *Game) RunNumberThenForEffect(b NumberBranch, chooser, source uuid.UUID, n int) error {
	if b.IsZero() {
		return nil
	}
	return g.runNumberBranchNow(b, chooser, source, n)
}

// apply sets the keyed fields of a number prompt about to be queued, and
// the cause it is queued under.
func (b NumberBranch) apply(c *PendingChoice, cause MoveCause) {
	c.PickThen = b.Then.key
	c.OptionCarry = copyUUIDs(b.Carry)
	c.PickWords = copyStrings(b.Words)
	c.PickNums = copyInts(b.Nums)
	c.Asked = copyAskedItem(b.Asked)
	// #4279: the cause is taken while the asking resolution is open. The
	// answer may come after a restore, when nothing is resolving.
	c.PayCauseController, c.PayCauseItem = cause.Controller, cause.Item
}

// numberFrameOf is the continuation behind a keyed number prompt, built
// from its data: the number answered, or "" when the chooser left (a
// payment was not made, and a number that is not paid is its floor). Nil
// for any other prompt, including a MayPayX's "choose X", which names
// pay-unless branches (mayPayXFrameOf).
func numberFrameOf(c *PendingChoice) *chooseValueFrame {
	if c == nil || c.Kind != PendingChoicePayAmount || c.MayPayX != nil || c.PickThen == "" {
		return nil
	}
	fallback := 0
	if c.PayAmount != nil && c.PayAmount.ResourceOrEnergy() == PayResourceNone {
		fallback = c.PayAmount.Min
	}
	key := c.PickThen
	base := NumberBranch{
		Carry: copyUUIDs(c.OptionCarry), Words: copyStrings(c.PickWords), Nums: copyInts(c.PickNums),
		Asked: copyAskedItem(c.Asked),
	}
	chooser, source := c.Chooser, c.Source
	assist, split := cloneAssistPrompt(c.Assist), cloneAttackSplit(c.AttackSplit)
	cause := carriedCauseOfChoice(c)
	return &chooseValueFrame{then: func(g *Game, value string) error {
		n, err := strconv.Atoi(value)
		if err != nil {
			n = fallback
		}
		r := base.answer(chooser, source, n)
		r.assist, r.split = cloneAssistPrompt(assist), cloneAttackSplit(split)
		return g.runNumberBranch(key, r, cause)
	}}
}
