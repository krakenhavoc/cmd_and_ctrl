package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_number.go — a number chosen as an effect resolves or as a
// permanent enters (ADR 0129's amendment of 2026-10-09, #1941). The
// pay_amount prompt underneath is ADR 0129 §3's; see
// game/choose_number.go for the rules and the shape.
//
// Three builders, one per printed shape:
//
//   - ChooseNumber: "an amount of damage of your choice" (Volcano
//     Hellion). A number chosen and not paid (CR 608.2d), with or
//     without a ceiling.
//   - PayLifeAmount: "you may pay any amount of life. If you do, …"
//     (Necrodominance). A cost paid on resolution (CR 118.12), capped at
//     the payer's life total (CR 119.4).
//   - PayAnyAmountOfLifeAsEnters: "As this enters, pay any amount of
//     life" (Phyrexian Processor), for Spec.AsEnters. The amount is
//     stored on the permanent; read it back with LifePaidAsEntered.
//
// Each Then runs exactly once, with the number, also when nothing could
// be asked: with 0 for a payment, and with Min for a number. A Goal is
// the card's own threshold, computed as the prompt is asked: the bot
// answers it when it can afford to, and a person's stepper opens on it.

// ChooseNumber asks the chooser (default: the controller) for a number
// between Min and Max, or of Min or more with NoMax.
type ChooseNumber struct {
	Chooser  uuid.UUID
	Min, Max int
	NoMax    bool
	Question string
	// Unit is what each point buys (game.PayAmountDamage, …).
	Unit string
	// SelfDamage marks a number that is also dealt to the chooser
	// ("to you and target creature"), so a bot prices it in life.
	SelfDamage bool
	Goal       func(ctx *Context) int
	// Marks are further meaningful numbers the bot is offered.
	Marks func(ctx *Context) []int
	Then  func(ctx *Context, n int) error
}

func (c ChooseNumber) Apply(ctx *Context) error {
	var marks []int
	if c.Marks != nil {
		marks = c.Marks(ctx)
	}
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:    energyChooser(ctx, c.Chooser),
		Source:     ctx.Source(),
		Resource:   game.PayResourceNone,
		Min:        c.Min,
		Max:        c.Max,
		NoMax:      c.NoMax,
		Goal:       goalOf(ctx, c.Goal),
		Marks:      marks,
		Unit:       c.Unit,
		SelfDamage: c.SelfDamage,
		Question:   c.Question,
		Then:       amountThen(ctx.Item, c.Then),
	})
}

// PayLifeAmount is "you may pay any amount of life" as an effect
// resolves. Then runs with the life paid: 0 when the payer declined,
// had no life to pay or left.
type PayLifeAmount struct {
	Chooser  uuid.UUID
	Question string
	Unit     string
	Goal     func(ctx *Context) int
	Then     func(ctx *Context, paid int) error
}

func (p PayLifeAmount) Apply(ctx *Context) error {
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  energyChooser(ctx, p.Chooser),
		Source:   ctx.Source(),
		Resource: game.PayResourceLife,
		Goal:     goalOf(ctx, p.Goal),
		Unit:     p.Unit,
		Question: p.Question,
		Then:     amountThen(ctx.Item, p.Then),
	})
}

// PayAnyAmountOfLifeAsEnters builds the Spec.AsEnters for "As this
// permanent enters, pay any amount of life" (CR 614.1c). The payer is
// the entering permanent's controller, the life paid is stored on the
// permanent (game.Card.ChosenNumber), and `goal` is what the card would
// have its controller pay given their life total (nil names none).
//
// The as-enters family's declared simplification applies: the prompt is
// queued from the AsEnters hook rather than by pausing the CR 614
// pipeline, so the permanent is briefly on the battlefield with nothing
// paid. The prompt blocks the table, so nothing can act in that window,
// and what reads the amount (an activated ability) cannot be activated
// in it.
func PayAnyAmountOfLifeAsEnters(label, unit string, goal func(life int) int) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		id := card.InstanceID
		g0 := 0
		if p := ctx.Game.PlayerByIDForEffect(card.Controller); p != nil && goal != nil {
			g0 = goal(p.Life)
		}
		return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
			Chooser:  card.Controller,
			Source:   id,
			Resource: game.PayResourceLife,
			Goal:     g0,
			Unit:     unit,
			Question: label + " — pay any amount of life",
			Then: func(g *game.Game, paid int) error {
				g.SetChosenNumberForEffect(id, paid)
				return nil
			},
		})
	}
}

// LifePaidAsEntered is "the life paid as this artifact entered", read
// as an ability of the permanent resolves: off the permanent while it is
// on the battlefield, and off its last-known information once it has
// left (CR 608.2h). Zero when there is no such permanent.
func LifePaidAsEntered(ctx *Context) int {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return 0
	}
	return info.ChosenNumber
}

// amountThen is the engine-side continuation of a number prompt asked
// while `item` resolves: `then` with a fresh Context on the game it is
// handed (an undo restores a clone), or nothing when there is no `then`.
// Shared by every builder here and by PayEnergyAmount.
func amountThen(item *game.StackItem, then func(ctx *Context, n int) error) func(g *game.Game, n int) error {
	return func(g *game.Game, n int) error {
		if then == nil {
			return nil
		}
		return then(NewContext(g, item), n)
	}
}

// goalOf evaluates a builder's Goal, or 0 when it has none.
func goalOf(ctx *Context, goal func(ctx *Context) int) int {
	if goal == nil {
		return 0
	}
	return goal(ctx)
}
