package game

import (
	"slices"

	"github.com/google/uuid"
)

// choose_number.go — a number chosen as an effect resolves, or as a
// permanent enters (ADR 0129's amendment of 2026-10-09, #1941).
//
// CR 608.2d: "If an effect of a spell or ability offers any choices
// other than choices already made as part of casting the spell,
// activating the ability, or otherwise putting the spell or ability on
// the stack, the player announces these while applying the effect."
// Some of those choices are a number with no printed list:
//
//   - "it deals an amount of damage of your choice to you and target
//     creature" (Volcano Hellion): a number chosen and not paid, with no
//     ceiling. CR 107.1b rules out a negative one and nothing else.
//   - "you may pay any amount of life. If you do, draw that many cards"
//     (Necrodominance): a number paid in life, which CR 119.4 caps at
//     the payer's life total. CR 118.12 makes it a cost paid on
//     resolution, so it may always be declined (paying 0).
//   - "As this artifact enters, pay any amount of life" (Phyrexian
//     Processor): the same payment made as a permanent enters
//     (CR 614.1c, 614.12a), and stored on it (Card.ChosenNumber) for
//     the ability that reads "the life paid as this artifact entered".
//
// All three are the pay_amount prompt ADR 0129 §3 built for "pay any
// amount of {E}". The owner chose that prompt over a long option list
// knowing it "can later serve 'pay any amount of life' or mana" (owner
// decision 3), so this file widens it rather than adding a kind:
// PayAmountPrompt.Resource says what each point costs, NoMax lifts the
// ceiling and Marks names the numbers worth offering a bot. The gate,
// the departure rule, the enumerator, the wire field, the client's
// stepper and the dispatcher's `amount` answer are the ones energy
// already has.
//
// The continuation is a registered branch (ChooseNumber.Branch, #4294),
// so the prompt holds no closure and a table holding one is a restore
// point. The closure frame it once rode is gone (#4361).
//
// A payment in MANA ("you may pay {X}{R}", Tilonalli's Summoner, #2727)
// takes exactly that shape (pay_x_on_resolution.go): the number is
// chosen here with PayResourceNone and a ceiling of what the payer could
// pay, and its branch queues the existing pay_unless prompt for the cost with
// X filled in, so the mana is paid through the one path that pays mana
// for an effect.

// ChooseNumber is what QueueChooseNumberForEffect is handed.
type ChooseNumber struct {
	// Chooser answers the prompt. Source attributes it to a card.
	Chooser, Source uuid.UUID
	// Resource is PayResourceLife or PayResourceNone. Energy has its
	// own entry point, QueuePayEnergyAmountForEffect.
	Resource string
	// Min is the smallest number: 0 for "any amount".
	Min int
	// Max is the printed ceiling of a number that is not paid. A life
	// payment's ceiling is the payer's life total (CR 119.4) and Max is
	// ignored for it; NoMax is a number with no ceiling at all.
	Max   int
	NoMax bool
	// Goal is the number that reaches what the card is for, or 0: the
	// target's lethal damage for Volcano Hellion. A bot answers it when
	// it can afford to; a person's stepper opens on it.
	Goal int
	// Marks are further meaningful numbers the enumerator offers.
	Marks []int
	// Unit is what each point buys (PayAmountDamage, PayAmountCards, …).
	Unit string
	// SelfDamage marks a number that is also dealt to the chooser.
	SelfDamage bool
	// Question is the prompt's header.
	Question string
	// Branch is the rest of the card, a REGISTERED branch
	// (RegisterNumberThen, #4294) handed the number as
	// NumberAnswer.Number once it is chosen (and paid). It runs exactly
	// once, also when there was nothing to ask and when the chooser left:
	// with 0 for a payment, which was not made, and with Min for a number
	// that is not paid. Zero is a number nothing follows. The prompt holds
	// no closure, so a table waiting on the chooser is a restore point
	// (the closure form, Then, is gone: #4361).
	Branch NumberBranch
	// split is the engine's attack split's state (attack_split.go).
	split *AttackSplit
}

// QueueChooseNumberForEffect asks p.Chooser for a number (CR 608.2d).
// When there is nothing to choose — a life payment whose payer has no
// life to pay, or whose life total can't change (CR 119.8) — nobody is
// asked and Branch runs with 0. A chooser who has left is not asked either.
//
// Caller must hold g.mu.
func (g *Game) QueueChooseNumberForEffect(p ChooseNumber) error {
	floor := max(p.Min, 0)
	fallback := 0
	if p.Resource == PayResourceNone {
		fallback = floor
	}
	// run is the branch when nobody is asked; an asked prompt runs it from
	// its data (numberFrameOf).
	run := func(g *Game, n int) error {
		if p.Branch.IsZero() {
			return nil
		}
		// A split's state rides the answer, also when nobody is asked.
		r := p.Branch.answer(p.Chooser, p.Source, n)
		r.split = cloneAttackSplit(p.split)
		return g.runNumberBranch(p.Branch.Then.key, r, carriedCause{})
	}
	pl := g.playerByIDLocked(p.Chooser)
	if pl == nil || pl.Eliminated {
		return run(g, fallback)
	}
	ceiling := p.Max
	switch {
	case p.Resource == PayResourceLife:
		ceiling = g.lifePayableLocked(pl)
		p.NoMax = false
	case p.NoMax:
		ceiling = PayAmountNoMaxCeiling
	}
	if ceiling <= 0 || ceiling < floor {
		// Nothing to choose: no life to pay, or a ceiling under the
		// floor (CR 118.3).
		return run(g, fallback)
	}
	goal := p.Goal
	if goal < floor || goal > ceiling {
		goal = 0
	}
	var marks []int
	for _, m := range p.Marks {
		if m >= floor && m <= ceiling && !slices.Contains(marks, m) {
			marks = append(marks, m)
		}
	}
	slices.Sort(marks)
	unit := p.Unit
	if unit == "" {
		unit = PayAmountOther
	}
	resource := p.Resource
	if resource == "" {
		resource = PayResourceNone
	}
	prompt := PendingChoice{
		Kind:    PendingChoicePayAmount,
		Chooser: p.Chooser,
		Count:   1,
		Source:  p.Source,
		Reason:  p.Question,
		PayAmount: &PayAmountPrompt{
			Min: floor, Max: ceiling, Goal: goal, Unit: unit,
			Resource: resource, NoMax: p.NoMax, Marks: marks, SelfDamage: p.SelfDamage,
		},
	}
	if !p.Branch.IsZero() {
		// #4294: the keyed form, a restore point.
		p.Branch.apply(&prompt, g.resolutionCauseLocked())
		prompt.AttackSplit = cloneAttackSplit(p.split)
	}
	// #4361: with no branch nothing follows the answer, so the prompt
	// carries no frame and is plain data too.
	id := g.QueueChoiceForEffect(prompt)
	if id == uuid.Nil {
		return run(g, fallback)
	}
	return nil
}

// lifePayableLocked is the most life `p` can pay now: their life total
// (CR 119.4), or 0 when it is not positive or can't change (CR 119.8).
//
// Caller must hold g.mu.
func (g *Game) lifePayableLocked(p *Player) int {
	if p == nil || p.Life <= 0 || !g.CanPayLifeLocked(p, 1) {
		return 0
	}
	return p.Life
}

// SetChosenNumberForEffect stores n on the battlefield permanent `id`
// as its Card.ChosenNumber: the number chosen as it entered. A
// permanent that has left stores nothing. Not a characteristic, so no
// layer version moves.
//
// Caller must hold g.mu.
func (g *Game) SetChosenNumberForEffect(id uuid.UUID, n int) {
	if c := findBattlefieldCard(g, id); c != nil {
		c.ChosenNumber = max(n, 0)
	}
}

// ChosenNumberOf is the number stored on the battlefield permanent `id`
// as it entered, or 0 when it is not on the battlefield or chose none.
// Read-only; safe under either lock.
func (g *Game) ChosenNumberOf(id uuid.UUID) int {
	if c := findBattlefieldCard(g, id); c != nil {
		return c.ChosenNumber
	}
	return 0
}
