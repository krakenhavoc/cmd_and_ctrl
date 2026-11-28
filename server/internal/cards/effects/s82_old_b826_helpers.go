package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b826_helpers.go — Homelands, batch 826 (S82 older sets, #5635).

// OldB826DrawUpToTwo is Truce's "each player may draw up to two cards. For
// each card less than two a player draws this way, that player gains 2 life",
// as a spell body. It is Temporary Truce's shape under its own keys: each
// player, the active player first and then in turn order, is asked for a
// number from 0 to two (a draw limit caps it, CR 121.2b), draws that many and
// then gains 2 life for each card short of two, counted from the draws that
// really arrived. keyBase names the registered continuations ("truce"), and
// name the prompt. Call it once, from the card's init.
func OldB826DrawUpToTwo(keyBase, name string) func(*game.StackItem, *Context) error {
	record := keyBase + "/chosen"
	baseline := keyBase + "/draws-before"
	var chosen game.NumberThen
	var ask func(ctx *Context, order []uuid.UUID) error
	rest := RegisterThenSteps("then-steps/"+keyBase+"-rest",
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(record)
			if len(r.IDs) == 0 {
				return nil
			}
			ctx.RecordStep(baseline, nil, GapB34DrawEvents(ctx.Game, r.IDs[0]))
			return DrawCards{Player: r.IDs[0], N: r.N()}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(record)
			if len(r.IDs) == 0 {
				return nil
			}
			drawn := GapB34DrawEvents(ctx.Game, r.IDs[0]) - ctx.StepRecord(baseline).N()
			if short := 2 - drawn; short > 0 {
				return GainLife{Player: r.IDs[0], Amount: 2 * short}.Apply(ctx)
			}
			return nil
		}),
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(record)
			if len(r.IDs) < 2 {
				return nil
			}
			return ask(ctx, r.IDs[1:])
		}),
	)
	chosen = NumberThen("number/"+keyBase+"-draw-up-to-two", func(ctx *Context, r game.NumberAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		ctx.RecordStep(record, r.Carry, r.Number)
		return rest.Run(ctx)
	})
	ask = func(ctx *Context, order []uuid.UUID) error {
		for len(order) > 0 && ctx.PlayerByID(order[0]) == nil {
			order = order[1:]
		}
		if len(order) == 0 {
			return nil
		}
		who := order[0]
		return ChooseNumber{
			Chooser:  who,
			Min:      0,
			Max:      ctx.Game.DrawChoiceCeilingForEffect(who, 2),
			Question: name + " — draw up to two cards (gain 2 life for each card less than two)",
			Unit:     game.PayAmountCards,
			Goal:     func(*Context) int { return 2 },
			Then:     chosen,
			Carry:    order,
		}.Apply(ctx)
	}
	return func(_ *game.StackItem, ctx *Context) error {
		return ask(ctx, GapB34TableInTurnOrder(ctx))
	}
}

// OldB826BlocksOrBecomesBlockedByWhite is Ghost Hounds' trigger condition: the
// source is one half of a block pair (CR 509.3d: one EventBlock per
// blocker-attacker pair) whose other creature is white.
func OldB826BlocksOrBecomesBlockedByWhite(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	foe := SkewerSlingerFoe(ev, source.InstanceID)
	if foe == uuid.Nil {
		return false
	}
	other, ok := g.LookupCardForEffect(foe)
	return ok && other.IsCreature() && other.HasColorWord("W")
}

// OldB826CountSubtypes counts the permanents `controller` controls that have
// any of the subtypes, each once.
func OldB826CountSubtypes(g *game.Game, controller uuid.UUID, subtypes ...string) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != controller {
			continue
		}
		for _, s := range subtypes {
			if c.HasSubtype(s) {
				n++
				break
			}
		}
	}
	return n
}

// OldB826GreenCreatures counts the green creatures on the battlefield,
// whoever controls them.
func OldB826GreenCreatures(g *game.Game) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsCreature() && c.HasColorWord("G") {
			n++
		}
	}
	return n
}
