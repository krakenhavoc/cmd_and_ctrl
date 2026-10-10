package effects

import (
	"slices"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_pay_x.go — "you may pay {X}{R}. If you do, …" as an ability
// resolves (#2727, ADR 0129's amendment of 2026-10-09), and "create X
// tokens that are tapped and attacking", each attacking what its
// controller chooses (CR 508.4). The engine half of the payment is
// game/pay_x_on_resolution.go.

// MayPayX is "you may pay <Cost with one {X}>. If you do, <OnPay with
// X>". X is chosen as the ability resolves (CR 608.2d), from 0 to the
// most the chooser can pay right now, and then the cost with X settled
// is offered as an ordinary "you may pay" (MayPay's prompt and payment
// path). Like MayPay, it only queues: anything the card does after the
// payment lives in OnPay.
type MayPayX struct {
	// Chooser is who picks X and pays: the controller when zero.
	Chooser uuid.UUID
	// Cost is the printed cost: "{X}{R}".
	Cost string
	// Label names the card in the prompts.
	Label string
	// Buys says what X buys in the payment's header, with "X" standing
	// for the number: "create X tapped and attacking Elementals".
	Buys string
	// Goal is the X a bot answers, given the most the chooser can pay.
	// Nil is 0: the bot keeps its mana.
	Goal func(ctx *Context, ceiling int) int
	// Unit is what each point of X buys (game.PayAmountDamage, …).
	Unit string
	// InThisStep holds the payment in the current step: set it when X
	// buys something about the step in progress (tokens that enter
	// attacking).
	InThisStep bool
	OnPay      func(ctx *Context, x int) error
}

func (m MayPayX) Apply(ctx *Context) error {
	item := ctx.Item
	var goal func(int) int
	if m.Goal != nil {
		goal = func(ceiling int) int { return m.Goal(ctx, ceiling) }
	}
	onPay := m.OnPay
	return ctx.Game.QueueMayPayXForEffect(game.MayPayX{
		Chooser:    energyChooser(ctx, m.Chooser),
		Source:     ctx.Source(),
		Cost:       m.Cost,
		Label:      m.Label,
		Buys:       m.Buys,
		Goal:       goal,
		Unit:       m.Unit,
		InThisStep: m.InThisStep,
		OnPay: func(g *game.Game, x int) error {
			if onPay == nil {
				return nil
			}
			return onPay(NewContext(g, item), x)
		},
	})
}

// XAsHighAsYouCan is a MayPayX Goal: every point the chooser can pay.
func XAsHighAsYouCan(_ *Context, ceiling int) int { return ceiling }

// CreateTokensAttackingYourChoice is "create N <Template> tokens that
// are tapped and attacking" where the card names no player: CR 508.4
// has the tokens' controller choose what each one attacks, among the
// players, planeswalkers and battles they could attack
// (Game.AttackTargetsForEffect).
//
// With one thing to attack there is nothing to ask. With more, the
// controller is asked how many of the tokens left attack each one in
// turn, Prefer first (what the creature whose trigger this is
// attacks), and the last takes the rest — at most one question per
// attackable thing but the last, and none once every token is placed.
// A bot sends them all at Prefer. The tokens are then made by one
// creation (game.TokenGroup.Attacking carries each share), so a token
// doubler doubles each share and a "whenever one or more tokens are
// created" trigger fires once.
//
// Then runs with every token made, also when there is nothing to
// attack (CR 508.4a: they are made, not attacking).
type CreateTokensAttackingYourChoice struct {
	Controller uuid.UUID
	Template   game.Card
	N          int
	Tapped     bool
	Prefer     uuid.UUID
	// Label names the card in the questions.
	Label string
	Then  func(ctx *Context, created []uuid.UUID) error
}

func (c CreateTokensAttackingYourChoice) Apply(ctx *Context) error {
	if c.N <= 0 {
		return nil
	}
	controller := c.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	var targets []uuid.UUID
	for _, ref := range ctx.Game.AttackTargetsForEffect(controller) {
		targets = append(targets, ref.ID)
	}
	if i := slices.Index(targets, c.Prefer); i > 0 {
		targets = append([]uuid.UUID{c.Prefer}, slices.Delete(targets, i, i+1)...)
	}
	if len(targets) <= 1 {
		return c.create(ctx, controller, targets, []int{c.N})
	}
	return c.ask(ctx, controller, targets, nil, c.N)
}

// ask places the tokens still unplaced: `shares` holds how many attack
// each of targets[:len(shares)], and `left` are still to go.
func (c CreateTokensAttackingYourChoice) ask(ctx *Context, controller uuid.UUID, targets []uuid.UUID, shares []int, left int) error {
	i := len(shares)
	if left == 0 || i == len(targets)-1 {
		return c.create(ctx, controller, targets, append(slices.Clone(shares), left))
	}
	goal := 0
	if i == 0 {
		goal = left
	}
	item := ctx.Item
	return ctx.Game.QueueChooseNumberForEffect(game.ChooseNumber{
		Chooser:  controller,
		Source:   ctx.Source(),
		Resource: game.PayResourceNone,
		Max:      left,
		Goal:     goal,
		Question: c.Label + " — how many of the " + strconv.Itoa(left) + " tokens attack " + attackTargetName(ctx.Game, targets[i]) + "?",
		Then: func(g *game.Game, n int) error {
			n = min(max(n, 0), left)
			return c.ask(NewContext(g, item), controller, targets, append(slices.Clone(shares), n), left-n)
		},
	})
}

// create makes the tokens in one creation, shares[i] of them attacking
// targets[i] (or none attacking when there is no target).
func (c CreateTokensAttackingYourChoice) create(ctx *Context, controller uuid.UUID, targets []uuid.UUID, shares []int) error {
	var groups []game.TokenGroup
	for i, n := range shares {
		if n <= 0 {
			continue
		}
		grp := game.TokenGroup{Template: c.Template, Count: n, Entry: game.TokenEntryOptions{Tapped: c.Tapped}}
		if i < len(targets) {
			grp.Attacking = targets[i]
		}
		groups = append(groups, grp)
	}
	item := ctx.Item
	then := c.Then
	return ctx.Game.CreateTokensThenForEffect(game.TokenCreation{
		Controller: controller,
		Groups:     groups,
		Source:     ctx.Source(),
	}, func(g *game.Game, created []uuid.UUID) error {
		if then == nil {
			return nil
		}
		return then(NewContext(g, item), created)
	})
}

// attackTargetName is a player's or a permanent's name for a question.
func attackTargetName(g *game.Game, id uuid.UUID) string {
	if p := g.PlayerByIDForEffect(id); p != nil && p.Name != "" {
		return p.Name
	}
	if c, ok := g.LookupCardForEffect(id); ok && c.Name != "" {
		return c.Name
	}
	return "that player"
}

// ExileUnlessCitysBlessing schedules "At the beginning of the next end
// step, exile those tokens unless you have the city's blessing"
// (Tilonalli's Summoner). The blessing is read as the delayed trigger
// resolves, so a player who gets it before that end step keeps the
// tokens.
func ExileUnlessCitysBlessing(ctx *Context, label string, cards []uuid.UUID) error {
	if len(cards) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: label,
		Cards: cards,
		Body:  exileListedUnlessCitysBlessingBody,
	}.Apply(ctx)
}

// exileListedUnlessCitysBlessing is the body: nothing when the delayed
// trigger's controller has the city's blessing, else exile every listed
// card still on the battlefield.
func exileListedUnlessCitysBlessing(g *game.Game, item *game.StackItem) error {
	if YouHaveTheCitysBlessing(g, item.Controller) {
		return nil
	}
	return b06ExileListedCards(g, item)
}
