package effects

import (
	"fmt"

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
//
// Then is a REGISTERED branch (NumberThen, #4294), handed the number as
// r.Number and what it needs as data: the stack item that asked
// (ctx.Item, rebuilt from game.AskedItem), the card's own Carry, Words
// and Nums. No prompt here holds a closure, so a table waiting on a
// number is a restore point. Share a key where the consequence is the
// same ("deal that much damage to the target").

// NumberThen registers a card's keyed number branch (#4294): the body
// runs with the number as r.Number against the stack item that asked
// (ctx.Item), or, for a prompt with none, a stand-in carrying the
// controller and the source card. The asking effect's controller rides
// first on the prompt's Carry (numberBranchOf), and r.Carry is what the
// card put after it. The body captures nothing: everything it reads is
// the prompt's data or the game. The key is an on-disk identity, never
// renamed, never reused.
func NumberThen(key string, body func(ctx *Context, r game.NumberAnswer) error) game.NumberThen {
	return game.RegisterNumberThen(key, func(g *game.Game, r game.NumberAnswer) error {
		ctx, carry := keyedCarryContext(g, r.Chooser, r.Source, r.Carry)
		if r.Item != nil {
			ctx = NewContext(g, r.Item)
		}
		r.Carry = carry
		return body(ctx, r)
	})
}

// numberBranchOf is the branch a builder queues its prompt with: the
// registered key, the controller first on the Carry, the card's own data
// after it, and the asking item as data.
func numberBranchOf(ctx *Context, then game.NumberThen, carry []uuid.UUID, words []string, nums []int) (game.NumberBranch, error) {
	if then.Key() == "" {
		return game.NumberBranch{}, fmt.Errorf("%w: a number prompt with no Then", game.ErrInvalidParam)
	}
	return game.NumberBranch{
		Then:  then,
		Carry: append([]uuid.UUID{ctx.Controller()}, carry...),
		Words: words,
		Nums:  nums,
		Asked: game.NewAskedItem(ctx.Item),
	}, nil
}

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
	// Then is the branch the number is handed to; Carry, Words and Nums
	// are the plain data it is handed with it.
	Then  game.NumberThen
	Carry []uuid.UUID
	Words []string
	Nums  []int
}

func (c ChooseNumber) Apply(ctx *Context) error {
	var marks []int
	if c.Marks != nil {
		marks = c.Marks(ctx)
	}
	branch, err := numberBranchOf(ctx, c.Then, c.Carry, c.Words, c.Nums)
	if err != nil {
		return err
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
		Branch:     branch,
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
	Then     game.NumberThen
	Carry    []uuid.UUID
	Words    []string
	Nums     []int
}

func (p PayLifeAmount) Apply(ctx *Context) error {
	branch, err := numberBranchOf(ctx, p.Then, p.Carry, p.Words, p.Nums)
	if err != nil {
		return err
	}
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  energyChooser(ctx, p.Chooser),
		Source:   ctx.Source(),
		Resource: game.PayResourceLife,
		Goal:     goalOf(ctx, p.Goal),
		Unit:     p.Unit,
		Question: p.Question,
		Branch:   branch,
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
			// #4294: keyed, with no asking item (an "as it enters" choice
			// has none). The controller rides first on the Carry, as it
			// does for every keyed number prompt.
			Branch: game.NumberBranch{Then: storeChosenNumber, Carry: []uuid.UUID{card.Controller}},
		})
	}
}

// storeChosenNumber is the branch behind "As this enters, pay any amount
// of life": the life paid is stored on the entering permanent, which is
// the prompt's source.
var storeChosenNumber = NumberThen("number/store-the-life-paid-as-it-enters", func(ctx *Context, r game.NumberAnswer) error {
	ctx.Game.SetChosenNumberForEffect(r.Source, r.Number)
	return nil
})

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

// goalOf evaluates a builder's Goal, or 0 when it has none.
func goalOf(ctx *Context, goal func(ctx *Context) int) int {
	if goal == nil {
		return 0
	}
	return goal(ctx)
}
