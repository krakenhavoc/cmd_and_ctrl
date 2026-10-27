package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_if_dies_g3.go — shared bodies for the third batch of ADR 0108
// PR 1 cards (#1886, #1887): the modal bullets, the activated
// abilities and the multi-amount damage that exile_if_dies.go's spell
// bodies do not cover. Append-only: add a builder, never change what
// one means.

// damageModesTargetExileIfItDies is a modal bullet's "~ deals N damage
// to target creature [or planeswalker]. If that creature [or
// planeswalker] would die this turn, exile it instead." (Agate
// Assault, Suplex, Pinecone Strike, Brutal Expulsion). It reads only
// this occurrence's own target. The replacement is the spell's, so it
// is registered whether or not the damage was dealt (ADR 0108 §1
// decision 3, the Disintegrate ruling).
func damageModesTargetExileIfItDies(n int) func(item *game.StackItem, ctx *Context, occ int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok {
			return nil
		}
		if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
			return err
		}
		return ExileIfItWouldDieThisTurn{Target: t.ID}.Apply(ctx)
	}
}

// returnModesSpellOrPermanentToHand is a modal bullet's "Return target
// spell or <permanent> to its owner's hand" (Brutal Expulsion): a spell
// is returned from the stack without being countered (ReturnSpellToHand,
// Venser's verb), a permanent is bounced. Reads only this occurrence's
// own target, re-checked as it resolves (CR 608.2b).
func returnModesSpellOrPermanentToHand(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneStack {
		return ReturnSpellToHand{StackID: t.ID}.Apply(ctx)
	}
	return BounceTheModesTarget(item, ctx, occ)
}

// abilityBody is a spell-shaped body (item, ctx) as an activated
// ability's Effect, so an ability that says what a spell says (Jaya
// Ballard's Incinerate, Nine-Ringed Bo's Lava Coil) uses the same
// shared body as the spell.
func abilityBody(body func(item *game.StackItem, ctx *Context) error) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return body(item, NewContext(g, item))
	}
}

// damageStep is one "N damage to <recipient>" of a sentence that deals
// different amounts to different recipients (Serpentine Spike).
type damageStep struct {
	Target uuid.UUID
	Amount int
}

// dealDamageStepsThen deals each step's damage in order, each after the
// one before has settled (a CR 616 prompt included), with `then` told
// each recipient and what it was dealt — DealDamageToEachThen for a
// sentence whose amounts differ. A recipient that has gone is skipped
// and the rest are still dealt their damage.
func dealDamageStepsThen(ctx *Context, steps []damageStep, then RecipientThen) error {
	// One printed sentence is one damage instance (CR 615.8, ADR 0108
	// PR 0): every step joins the scope opened here.
	return ctx.Game.DamageInstanceForEffect(func() error {
		return dealDamageStepChain(ctx, steps, then)
	})
}

// dealDamageStepChain deals the head step and continues with the rest
// from its continuation, once the head's damage has settled. The chain is
// data (#4253): the rest of the steps ride damageStepsNext's tail, their
// recipients as its Carry and their amounts as its Nums, and the
// per-recipient rider as its Next.
func dealDamageStepChain(ctx *Context, steps []damageStep, then RecipientThen) error {
	return dealDamageStepsFrom(ctx.Game, ctx.Tail(damageStepsNext), steps, then)
}

// dealDamageStepsFrom deals the head of `steps` with `step`, carrying the
// rest, as its total's tail.
func dealDamageStepsFrom(g *game.Game, step game.KeyedTail, steps []damageStep, then RecipientThen) error {
	if len(steps) == 0 {
		return nil
	}
	head, rest := steps[0], steps[1:]
	step.Carry, step.Nums, step.Next = nil, nil, nil
	for _, s := range rest {
		step.Carry = append(step.Carry, s.Target)
		step.Nums = append(step.Nums, s.Amount)
	}
	if !then.IsZero() {
		each := then
		step.Next = &each
	}
	return g.DealDamageEachEachKeyedForEffect(step.Source, []uuid.UUID{head.Target}, head.Amount, then, step)
}

// damageStepsNext is one link of dealDamageStepChain, run with a step's
// total once its damage has settled: the next step is dealt. An on-disk
// key.
var damageStepsNext = game.RegisterTail("damage-steps/next", func(g *game.Game, o game.TailOutcome) error {
	var steps []damageStep
	for i, id := range o.Tail.Carry {
		if i < len(o.Tail.Nums) {
			steps = append(steps, damageStep{Target: id, Amount: o.Tail.Nums[i]})
		}
	}
	var then RecipientThen
	if o.Tail.Next != nil {
		then = *o.Tail.Next
	}
	return dealDamageStepsFrom(g, o.Tail, steps, then)
})

// damageEachMatchingAndEachPlayerThen is "~ deals N damage to each
// <creature predicate> and each player" with `then` told each
// recipient (Flamebreak). The creatures are the ones matching now, in
// battlefield order, then every player still in the game in seat
// order; one batch, so each is dealt its damage after the one before
// has settled.
func damageEachMatchingAndEachPlayerThen(ctx *Context, match CardPredicate, n int, then RecipientThen) error {
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, match) {
		ids = append(ids, c.InstanceID)
	}
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		ids = append(ids, p.ID)
	}
	return DealDamageToEachThen(ctx, ids, n, then)
}
