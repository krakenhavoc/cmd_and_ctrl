package game

import (
	"errors"
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
// A PayResourceMana that charged the pool from the stepper was the
// alternative the amendment named. It would have needed its own
// payment, its own auto-tap and its own Always / Never row for a
// question pay_unless already answers.

// MayPayX is what QueueMayPayXForEffect is handed.
type MayPayX struct {
	// Chooser answers both prompts and pays. Source attributes them.
	Chooser, Source uuid.UUID
	// Cost is the printed cost with exactly one {X}: "{X}{R}".
	Cost string
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
	// (QueueMayPayInThisStepForEffect): set it when what X buys is
	// about the step in progress — tokens that enter attacking.
	InThisStep bool
	// OnPay runs with X once the cost has been paid. It does not run
	// when the chooser declines, cannot fund the payment, or has left.
	OnPay func(g *Game, x int) error
}

// QueueMayPayXForEffect asks p.Chooser for X and then whether to pay
// p.Cost with X settled (see the file comment).
//
// Caller must hold g.mu.
func (g *Game) QueueMayPayXForEffect(p MayPayX) error {
	cost, err := ParseCost(p.Cost)
	if err == nil && cost.XSlots != 1 {
		err = errMayPayXNoX
	}
	if err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   p.Source,
			ErrorMsg: "may-pay-x: cost " + p.Cost + ": " + err.Error(),
		})
		return nil
	}
	pl := g.playerByIDLocked(p.Chooser)
	if pl == nil || pl.Eliminated {
		return nil
	}
	ceiling := g.largestPayableXLocked(pl, cost, p.Source)
	if ceiling < 0 {
		// Not even X = 0 can be paid: "Pay" is not an answer (CR
		// 118.3), and "Don't pay" is the only one left.
		return nil
	}
	goal := 0
	if p.Goal != nil {
		goal = p.Goal(ceiling)
	}
	onPay := p.OnPay
	return g.QueueChooseNumberForEffect(ChooseNumber{
		Chooser:  p.Chooser,
		Source:   p.Source,
		Resource: PayResourceNone,
		Min:      0,
		Max:      ceiling,
		Goal:     goal,
		Unit:     p.Unit,
		Question: p.Label + " — choose X for " + p.Cost + " (you can pay up to " + strconv.Itoa(ceiling) + ")",
		Then: func(g *Game, x int) error {
			settled := cost.SettleX(x)
			run := func(g *Game) error {
				if onPay == nil {
					return nil
				}
				return onPay(g, x)
			}
			if settled.Empty() {
				// {X} at X = 0 costs nothing (CR 118.5).
				return run(g)
			}
			price := settled.String()
			question := p.Label + " — pay " + price
			if p.Buys != "" {
				question += " to " + strings.ReplaceAll(p.Buys, "X", strconv.Itoa(x))
			}
			question += "?"
			if p.InThisStep {
				return g.QueueMayPayInThisStepForEffect(p.Chooser, p.Source, price, question, run)
			}
			return g.QueueMayPayForEffect(p.Chooser, p.Source, price, question, run)
		},
	})
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
func (g *Game) largestPayableXLocked(p *Player, cost ParsedCost, source uuid.UUID) int {
	ok := func(x int) bool { return g.canPayCostLocked(p, cost.SettleX(x), source) }
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
