package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_payment.go — ADR 0129 §3: the card-side vocabulary for paying
// energy while a spell or ability resolves. The engine half, and the
// rules (CR 118.12, 118.12a, 107.14), are in
// server/internal/game/energy_payment.go.
//
//	MayPayEnergy      "you may pay {E}{E}. If you do, …"
//	PayEnergyUnless   "sacrifice it unless you pay {E}"
//	PayEnergyOrElse   "pay {E}{E}. If you can't, …" (mandatory)
//	PayEnergyAmount   "you may pay any amount of {E}" / "one or more {E}"
//
// Like MayPay, the first two and the last only QUEUE a prompt: the
// calling effect has finished by the time the player answers, so what
// the card does after the payment lives inside OnPay / OnDecline / Then,
// never after Apply returns. Each branch gets a fresh Context bound to
// the same stack item.
//
// Every prompt here holds the table in the step it was asked in
// (game.EnergyPayment.InThisStep): an attack trigger's +1/+1 counter has
// to land before combat damage, and an upkeep draw before the draw step.

// MayPayEnergy is "you may pay N {E}. If you do, <OnPay>". Chooser
// defaults to the controller of the resolving spell or ability.
type MayPayEnergy struct {
	Chooser  uuid.UUID
	N        int
	Question string
	OnPay    func(ctx *Context) error
}

func (p MayPayEnergy) Apply(ctx *Context) error {
	return ctx.Game.QueueEnergyPaymentForEffect(game.EnergyPayment{
		Chooser:    energyChooser(ctx, p.Chooser),
		Source:     ctx.Source(),
		N:          p.N,
		Question:   p.Question,
		InThisStep: true,
		// The same adapter the decline uses: a fresh Context bound to
		// the same stack item, built when the branch runs.
		OnPay: declineAgainstTheSameItem(ctx.Item, p.OnPay),
	})
}

// PayEnergyUnless is "<OnDecline> unless you pay N {E}" (CR 118.12a).
// OnDecline runs on "no", on a "yes" the chooser cannot fund, and when
// the chooser has left the game (CR 800.4f).
type PayEnergyUnless struct {
	Chooser   uuid.UUID
	N         int
	Question  string
	OnDecline func(ctx *Context) error
}

func (p PayEnergyUnless) Apply(ctx *Context) error {
	return ctx.Game.QueueEnergyPaymentForEffect(game.EnergyPayment{
		Chooser:    energyChooser(ctx, p.Chooser),
		Source:     ctx.Source(),
		N:          p.N,
		Question:   p.Question,
		InThisStep: true,
		OnDecline:  declineAgainstTheSameItem(ctx.Item, p.OnDecline),
	})
}

// PayEnergyOrElse is the mandatory "pay N {E}. If you can't, <OrElse>"
// (Greenbelt Rampager). There is no question: the controller pays when
// they have the energy, and OrElse runs when they don't (CR 118.3).
type PayEnergyOrElse struct {
	N      int
	OrElse func(ctx *Context) error
}

func (p PayEnergyOrElse) Apply(ctx *Context) error {
	item := ctx.Item
	orElse := p.OrElse
	return ctx.Game.PayEnergyIfAbleForEffect(ctx.Controller(), ctx.Source(), p.N, func(g *game.Game, paid bool) error {
		if paid || orElse == nil {
			return nil
		}
		return orElse(NewContext(g, item))
	})
}

// PayEnergyAmount is "you may pay any amount of {E}" (Min 0) or "you may
// pay one or more {E}" (Min 1): the pay_amount prompt, whose answer is
// handed to Then. Then always runs exactly once, with 0 when nothing was
// paid — when the chooser had no energy, declined or left — so the rest
// of the card ("Destroy all creatures") lives inside it.
//
// Goal is the card's own threshold, computed when the prompt is asked:
// the smallest amount that does what the card is for (Harnessed
// Lightning: the target's toughness). Nil, or a result of 0, names none;
// a negative result is "as much as you can" (game.PayEnergyAmount). The
// bot pays it and the stepper starts there. Unit is what one counter
// buys (game.PayAmountDamage, …).
type PayEnergyAmount struct {
	Chooser  uuid.UUID
	Min      int
	Question string
	Unit     string
	Goal     func(ctx *Context) int
	Then     func(ctx *Context, paid int) error
}

func (p PayEnergyAmount) Apply(ctx *Context) error {
	item := ctx.Item
	then := p.Then
	goal := 0
	if p.Goal != nil {
		goal = p.Goal(ctx)
	}
	return ctx.Game.QueuePayEnergyAmountForEffect(game.PayEnergyAmount{
		Chooser:  energyChooser(ctx, p.Chooser),
		Source:   ctx.Source(),
		Min:      p.Min,
		Goal:     goal,
		Unit:     p.Unit,
		Question: p.Question,
		Then: func(g *game.Game, paid int) error {
			if then == nil {
				return nil
			}
			return then(NewContext(g, item), paid)
		},
	})
}

// energyChooser is the payer: the one named, or the controller of the
// resolving spell or ability. Energy comes off the payer (CR 107.14).
func energyChooser(ctx *Context, named uuid.UUID) uuid.UUID {
	if named != uuid.Nil {
		return named
	}
	return ctx.Controller()
}

// AsMuchAsYouCan is a PayEnergyAmount.Goal for a card whose every
// counter buys more of what its controller wants (Rampaging
// Aetherhood's +1/+1 counters, Pia Nalaar's X/X Aetherjet).
func AsMuchAsYouCan(*Context) int { return -1 }
