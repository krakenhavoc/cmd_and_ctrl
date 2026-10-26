package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_pay_x.go — "you may pay {X}{R}. If you do, …" as an ability
// resolves (#2727, ADR 0129's amendment of 2026-10-09), and "create X
// tokens that are tapped and attacking", each attacking what its
// controller chooses (CR 508.4). The engine half of the payment is
// game/pay_x_on_resolution.go.

// MayPayX is "you may pay <Cost with one {X}>. If you do, <Then with
// X>". X is chosen as the ability resolves (CR 608.2d), from 0 to the
// most the chooser can pay right now, and then the cost with X settled
// is offered as an ordinary "you may pay" (MayPay's prompt and payment
// path). Like MayPay, it only queues: anything the card does after the
// payment lives in Then.
//
// Then is the paid half as REGISTERED branches (PayUnlessBranches, #4289),
// handed X as r.Number: neither prompt holds a closure, so a table waiting
// on X or on the payment is a restore point. The branch runs against a
// copy of the asking item rebuilt from data the prompts carry
// (game.AskedItem).
type MayPayX struct {
	// Chooser is who picks X and pays: the controller when zero.
	Chooser uuid.UUID
	// Cost is the printed cost: "{X}{R}".
	Cost string
	// Symbol is "pay any amount of {R}": the cost is X of that one
	// symbol. Set it instead of Cost (Leyline Tyrant).
	Symbol string
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
	Then       game.PayUnlessElse
	Carry      []uuid.UUID
}

func (m MayPayX) Apply(ctx *Context) error {
	if m.Then.Key() == "" {
		return fmt.Errorf("%w: MayPayX has no Then", game.ErrInvalidParam)
	}
	var goal func(int) int
	if m.Goal != nil {
		goal = func(ceiling int) int { return m.Goal(ctx, ceiling) }
	}
	return ctx.Game.QueueMayPayXForEffect(game.MayPayX{
		Chooser:    energyChooser(ctx, m.Chooser),
		Source:     ctx.Source(),
		Cost:       m.Cost,
		Symbol:     m.Symbol,
		Label:      m.Label,
		Buys:       m.Buys,
		Goal:       goal,
		Unit:       m.Unit,
		InThisStep: m.InThisStep,
		Then:       m.Then,
		Carry:      append([]uuid.UUID{ctx.Controller()}, m.Carry...),
		Asked:      game.NewAskedItem(ctx.Item),
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
// Tail runs with every token made, also when there is nothing to attack
// (CR 508.4a: they are made, not attacking). It is a registered tail
// (AttackingTokensThen, #4294) and the split between the things to attack
// is plain data (game.AttackSplit), so a table waiting on one of the
// questions is a restore point. A card whose rest reads the asking item
// itself (Calamity) puts it on the tail (Context.ItemTail).
//
// Template is a description in the effect's text, read through it as
// the tokens are made (Context.TextToken, #4005), unless Copy (copiable
// values, CR 707.2: Sandstorm Crasher, Calamity) or RuleDefined (a
// keyword's token: mobilize's Warriors) says it is no word of the text.
type CreateTokensAttackingYourChoice struct {
	Controller  uuid.UUID
	Template    game.Card
	Copy        bool
	RuleDefined bool
	N           int
	Tapped      bool
	Prefer      uuid.UUID
	// Label names the card in the questions.
	Label string
	// Tail is the rest of the card as data; the effect's controller and
	// source fill its Player and Source where the card left them zero.
	Tail game.KeyedTail
}

// AttackingTokensThen registers a CreateTokensAttackingYourChoice.Tail
// body under `key`, an on-disk identity (never renamed, never reused,
// ledgered "tail <key>"). The body runs against a Context bound to a
// stand-in for the asking item: its controller and its source card. It is
// handed the tokens that landed, and the tail it was recorded with (its
// Carry and N). Call it once, from a package-level var.
func AttackingTokensThen(key string, body func(ctx *Context, created []uuid.UUID, tail game.KeyedTail) error) game.TailRef {
	return game.RegisterTail(key, func(g *game.Game, o game.TailOutcome) error {
		return body(keyedContext(g, o.Tail.Player, o.Tail.Source), o.IDs, o.Tail)
	})
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
	tail := c.Tail
	if !tail.IsZero() {
		if tail.Player == uuid.Nil {
			tail.Player = ctx.Controller()
		}
		if tail.Source == uuid.Nil {
			tail.Source = ctx.Source()
		}
	}
	return ctx.Game.StartAttackSplitForEffect(game.AttackSplit{
		Controller: controller,
		Source:     ctx.Source(),
		Template:   c.template(ctx),
		Count:      c.N,
		Tapped:     c.Tapped,
		Targets:    targets,
		Label:      c.Label,
		Tail:       tail,
	})
}

// template is the token as it is made: read through the effect's text
// (Context.TextToken, #4005), unless Copy (copiable values, CR 707.2:
// Sandstorm Crasher, Calamity) or RuleDefined (a keyword's token:
// mobilize's Warriors) says it is no word of the text.
func (c CreateTokensAttackingYourChoice) template(ctx *Context) game.Card {
	switch {
	case c.Copy:
		return copiedToken(c.Template)
	case c.RuleDefined:
		return ruleDefinedToken(c.Template)
	}
	return ctx.TextToken(c.Template)
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
