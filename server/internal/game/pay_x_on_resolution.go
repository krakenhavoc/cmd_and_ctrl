package game

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// pay_x_on_resolution.go — "you may pay {X}{R}. If you do, …" asked as
// an ability resolves (#2727, ADR 0129's amendment of 2026-10-09, the
// part headed "Paying {X} as an ability resolves").
//
// Tilonalli's Summoner: "Whenever this creature attacks, you may pay
// {X}{R}. If you do, create X 1/1 red Elemental creature tokens that
// are tapped and attacking." CR 608.2d makes X a choice made while the
// ability resolves, and CR 118.12 makes {X}{R} a cost paid then, so
// neither can move to the trigger's announcement: the controller picks
// X and pays in the same breath, with whatever mana they have at that
// moment.
//
// It is two prompts the engine already has, one after the other, as
// #1941's amendment proposed:
//
//  1. The number. pay_amount with PayResourceNone (choose_number.go):
//     0 up to the largest X the chooser could pay right now, read
//     through the same payment path the second prompt uses, so the
//     ceiling never offers an X the payment then refuses.
//  2. The payment. pay_unless's "you may pay" form for the cost with X
//     settled ("{3}{R}"), so the mana is paid through the one path that
//     pays mana for an effect: the pool first, then the auto-tapper,
//     with Always / Never (ADR 0127), the strict-mana gate and the
//     bot's pay-unless move all unchanged.
//
// Splitting the choice from the payment changes nothing the rules can
// see: the number prompt blocks the table, so nothing happens between
// the two answers, and "Don't pay" on the second is the "you may" the
// card prints. A chooser who cannot pay even X = 0 (no {R}) is not asked
// at all (CR 118.3), and a cost that is empty at X = 0 ("you may pay
// {X}" with X = 0) is paid by doing nothing (CR 118.5), so the payment
// prompt is skipped and the rider runs with 0.
//
// Both prompts are plain data (#4289), so a table waiting on either is a
// restore point. The number prompt holds no frame: it carries a
// MayPayXClause (the cost and the words of the next question), the key of
// the card's branches (PickThen, RegisterPayUnlessBranches), the asking
// item (Asked) and the cause it was queued under, and its answer builds
// the frame from that data (chooseValueFrameOf). The payment prompt is a
// keyed may-pay that carries X as PendingChoice.PickNumber, handed to the
// branch as PayUnlessAnswer.Number.
//
// A PayResourceMana that charged the pool from the stepper was the
// alternative the amendment named. It would have needed its own
// payment, its own auto-tap and its own Always / Never row for a
// question pay_unless already answers.

// MayPayX is what QueueMayPayXForEffect is handed.
type MayPayX struct {
	// Chooser answers both prompts and pays. Source attributes them.
	Chooser, Source uuid.UUID
	// Cost is the printed cost with exactly one {X}: "{X}{R}". Leave it
	// empty when Symbol is set.
	Cost string
	// Symbol is "pay ANY AMOUNT of {R}" (Leyline Tyrant): the cost is X of
	// this one mana symbol, "{R}", and nothing else. Set instead of Cost.
	Symbol string
	// Label names the card in the prompts' headers ("Tilonalli's
	// Summoner").
	Label string
	// Buys says what X buys, for the headers: "X tapped and attacking
	// Elementals" reads "create 3 tapped and attacking Elementals" once
	// X is settled. Every "X" in it is replaced with the number.
	Buys string
	// Goal is the X the card is for, given the largest X the chooser
	// can pay: a bot answers it and a person's stepper opens on it. Nil
	// is 0.
	Goal func(ceiling int) int
	// Unit is what each point of X buys (PayAmountDamage, …).
	Unit string
	// InThisStep holds the payment prompt in the current step
	// (QueueMayPayKeyedForEffect's inThisStep): set it when what X buys is
	// about the step in progress — tokens that enter attacking.
	InThisStep bool
	// Then is what the card does once the cost has been paid, as
	// REGISTERED branches (RegisterPayUnlessBranches, #4289): the pay
	// branch is handed X as PayUnlessAnswer.Number, and runs against the
	// asking item rebuilt from Asked. It does not run when the chooser
	// declines, cannot fund the payment, or has left. Neither prompt
	// holds a closure, so a table waiting on X, or on the payment, is a
	// restore point.
	Then PayUnlessElse
	// Carry is the plain data the branches are handed, and Asked the
	// asking stack item as data (NewAskedItem).
	Carry []uuid.UUID
	Asked *AskedItem
}

// MayPayXClause is the part of a MayPayX the "choose X" prompt must keep
// to ask the question after it (PendingChoice.MayPayX, #4289): plain data,
// so the prompt is written to a restore point. The branches that run once
// the cost is paid are the prompt's PickThen.
type MayPayXClause struct {
	// Cost is the printed cost ("{X}{R}"); for a Symbol cost it is the
	// words only ("X {R}").
	Cost string `json:"cost"`
	// Symbol is the one mana symbol a "pay any amount" cost repeats.
	Symbol string `json:"symbol,omitempty"`
	// Label and Buys word the payment's header, as MayPayX's do.
	Label string `json:"label,omitempty"`
	Buys  string `json:"buys,omitempty"`
	// InThisStep holds the payment prompt in the step it is asked in.
	InThisStep bool `json:"inThisStep,omitempty"`
}

// copyMayPayXClause is a copy of a clause, nil for nil.
func copyMayPayXClause(c *MayPayXClause) *MayPayXClause {
	if c == nil {
		return nil
	}
	out := *c
	return &out
}

// copyPayAmountPrompt is a deep copy of a pay_amount payload, nil for nil.
func copyPayAmountPrompt(pa *PayAmountPrompt) *PayAmountPrompt {
	if pa == nil {
		return nil
	}
	out := *pa
	out.Marks = copyInts(pa.Marks)
	return &out
}

// settler is the cost with X settled, as a function of X: X copies of
// the one symbol for a "pay any amount" cost, else the printed cost with
// its {X} filled in. It fails on a cost with no single {X}.
func (c *MayPayXClause) settler() (func(x int) ParsedCost, error) {
	if c.Symbol != "" {
		_, err := ParseCost(c.Symbol)
		return func(x int) ParsedCost {
			out, _ := ParseCost(strings.Repeat(c.Symbol, x))
			return out
		}, err
	}
	cost, err := ParseCost(c.Cost)
	if err == nil && cost.XSlots != 1 {
		err = errMayPayXNoX
	}
	return func(x int) ParsedCost { return cost.SettleX(x) }, err
}

// QueueMayPayXForEffect asks p.Chooser for X and then whether to pay
// p.Cost with X settled (see the file comment). Both prompts are plain
// data (#4289): the first names the branches in p.Then and carries the
// clause, the second is a keyed may-pay that carries X.
//
// Caller must hold g.mu.
func (g *Game) QueueMayPayXForEffect(p MayPayX) error {
	if p.Then.key == "" {
		return fmt.Errorf("%w: may-pay-x with no branches", ErrInvalidParam)
	}
	clause := &MayPayXClause{Cost: p.Cost, Symbol: p.Symbol, Label: p.Label, Buys: p.Buys, InThisStep: p.InThisStep}
	if p.Symbol != "" {
		clause.Cost = "X " + p.Symbol
	}
	settle, err := clause.settler()
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   p.Source,
			ErrorMsg: "may-pay-x: cost " + clause.Cost + ": " + err.Error(),
		})
		return nil
	}
	pl := g.playerByIDLocked(p.Chooser)
	if pl == nil || pl.Eliminated {
		return nil
	}
	ceiling := g.largestPayableXLocked(pl, settle, p.Source)
	if ceiling < 0 {
		// Not even X = 0 can be paid: "Pay" is not an answer (CR
		// 118.3), and "Don't pay" is the only one left.
		return nil
	}
	if ceiling == 0 {
		// Nothing to choose (CR 118.3): X is 0.
		return g.mayPayXAfterNumberLocked(clause, p.Then.key, p.Chooser, p.Source, p.Carry, p.Asked, 0)
	}
	goal := 0
	if p.Goal != nil {
		goal = p.Goal(ceiling)
	}
	if goal < 0 || goal > ceiling {
		goal = 0
	}
	unit := p.Unit
	if unit == "" {
		unit = PayAmountOther
	}
	// #4279: the cause is taken while the asking resolution is open; the
	// payment that follows is queued under it.
	cause := carriedCause{MoveCause: g.resolutionCauseLocked(), Source: p.Source}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoicePayAmount,
		Chooser: p.Chooser,
		Count:   1,
		Source:  p.Source,
		Reason:  p.Label + " — choose X for " + clause.Cost + " (you can pay up to " + strconv.Itoa(ceiling) + ")",
		PayAmount: &PayAmountPrompt{
			Min: 0, Max: ceiling, Goal: goal, Unit: unit, Resource: PayResourceNone,
		},
		MayPayX:     clause,
		PickThen:    p.Then.key,
		OptionCarry: copyUUIDs(p.Carry),
		Asked:       copyAskedItem(p.Asked),

		PayCauseController: cause.Controller,
		PayCauseItem:       cause.Item,
	})
	if id == uuid.Nil {
		return g.mayPayXAfterNumberLocked(clause, p.Then.key, p.Chooser, p.Source, p.Carry, p.Asked, 0)
	}
	return nil
}

// mayPayXAfterNumberLocked is what follows the number: with X settled, the
// cost is empty ({X} at X = 0 costs nothing, CR 118.5) and the paid branch
// runs at once, or the keyed may-pay for it is asked. Everything it needs
// is the prompt's data.
//
// Caller must hold g.mu.
func (g *Game) mayPayXAfterNumberLocked(clause *MayPayXClause, key string, chooser, source uuid.UUID, carry []uuid.UUID, asked *AskedItem, x int) error {
	settle, err := clause.settler()
	if err != nil {
		return nil
	}
	settled := settle(x)
	if settled.Empty() {
		r := PayUnlessAnswer{Chooser: chooser, Source: source, Carry: copyUUIDs(carry), Item: asked.StackItem(), Number: x}
		cause := carriedCause{MoveCause: g.resolutionCauseLocked(), Source: source}
		return keyedPayUnlessBranch(key, r, true, cause)(g)
	}
	price := settled.String()
	question := clause.Label + " — pay " + price
	if clause.Buys != "" {
		question += " to " + strings.ReplaceAll(clause.Buys, "X", strconv.Itoa(x))
	}
	question += "?"
	return g.QueuePayUnlessPromptKeyedForEffect(PayUnlessKeyedPrompt{
		Chooser: chooser, Source: source, Cost: price, Question: question,
		Then: PayUnlessElse{key: key}, Carry: carry, InThisStep: clause.InThisStep, MayPay: true,
		Asked: asked, Number: x,
	})
}

// mayPayXFrameOf is the continuation behind a keyed MayPayX's "choose X"
// prompt, built from its data: the number answered (or "" when the chooser
// left, which is X = 0) settles the cost and asks the payment. Nil for any
// other prompt.
func mayPayXFrameOf(c *PendingChoice) *chooseValueFrame {
	if c == nil || c.Kind != PendingChoicePayAmount || c.MayPayX == nil || c.PickThen == "" {
		return nil
	}
	clause := copyMayPayXClause(c.MayPayX)
	key, chooser, source := c.PickThen, c.Chooser, c.Source
	carry, asked := copyUUIDs(c.OptionCarry), copyAskedItem(c.Asked)
	cause := carriedCauseOfChoice(c)
	return &chooseValueFrame{then: func(g *Game, value string) error {
		x, err := strconv.Atoi(value)
		if err != nil || x < 0 {
			x = 0
		}
		return g.withCarriedCauseLocked(cause, func() error {
			return g.mayPayXAfterNumberLocked(clause, key, chooser, source, carry, asked, x)
		})
	}}
}

// chooseValueFrameOf is the frame a resolution-time value prompt answers
// through, built from its data: a keyed MayPayX's "choose X", a keyed
// number prompt (number_then.go), or a keyed colour or creature-type
// prompt (value_then_keyed.go). Nil for a prompt with none: a stored
// colour or type, or a pay_amount nothing follows. No prompt holds a frame
// of its own since #4361.
func chooseValueFrameOf(c *PendingChoice) *chooseValueFrame {
	if c == nil {
		return nil
	}
	if f := mayPayXFrameOf(c); f != nil {
		return f
	}
	if f := valueTailFrameOf(c); f != nil {
		// #4253: a keyed colour or creature-type prompt (value_then_keyed.go).
		return f
	}
	return numberFrameOf(c)
}

// errMayPayXNoX is a MayPayX cost without exactly one {X}: a catalog
// mistake, reported as EventEffectError.
var errMayPayXNoX = errors.New("the cost must have exactly one {X}")

// largestPayableXLocked is the largest X for which `p` can pay `cost`
// right now, or -1 when not even X = 0 can be paid. Each probe is
// canPayCostLocked — the payment the pay_unless prompt will make, run
// on a throwaway clone — so the ceiling is exactly what the payment
// accepts. The cost is monotonic in X (each point adds one generic
// symbol), so the search gallops up and then halves: a few clones,
// not one per point.
//
// Caller must hold g.mu.
func (g *Game) largestPayableXLocked(p *Player, settle func(x int) ParsedCost, source uuid.UUID) int {
	ok := func(x int) bool { return g.canPayCostLocked(p, settle(x), source) }
	if !ok(0) {
		return -1
	}
	lo, hi := 0, 1
	for ok(hi) {
		lo = hi
		if hi >= PayAmountNoMaxCeiling {
			return PayAmountNoMaxCeiling
		}
		hi *= 2
	}
	for hi-lo > 1 {
		mid := lo + (hi-lo)/2
		if ok(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}
