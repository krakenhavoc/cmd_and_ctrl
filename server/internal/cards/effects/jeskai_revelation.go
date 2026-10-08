package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeskai Revelation — Instant {4}{U}{R}{W}:
//
//	"Return target spell or permanent to its owner's hand. Jeskai
//	 Revelation deals 4 damage to any target. Create two 1/1 white Monk
//	 creature tokens with prowess. Draw two cards. You gain 4 life."
//
// Two target clauses (ADR 0065 §1): slot 0 is Venser's "target spell
// or permanent", slot 1 is "any target". Each is re-checked on its own
// at resolution, so a bounce target that left in response still lets
// the damage, the Monks, the draw and the life happen (CR 608.2b), and
// only both targets going illegal counters the spell.
//
// The sentences run in printed order. Returning a spell uses
// ReturnSpellToHand, so a spell that can't be countered still goes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61b6b1d1-4350-41c1-ac47-835b2831f24a",
		Name:         "Jeskai Revelation",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetSpellOrPermanent("target spell or permanent", nil, nil),
			TargetAny(),
		),
		Purpose: game.Purpose{
			Draws:    2,
			Tokens:   2,
			LifeGain: 4,
			Targets:  game.ForTargets(DamageToTarget(1, 4)),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := jeskaiRevelationReturn(ctx, t.ID); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (CreateToken{Template: TokenCard("1/1 white Monk with prowess"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		},
	})
}

// jeskaiRevelationReturn is Venser's verb for one already-checked
// target: a spell goes back from the stack without being countered,
// a permanent is bounced.
func jeskaiRevelationReturn(ctx *Context, id uuid.UUID) error {
	z := ctx.Game.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneStack:
		return ReturnSpellToHand{StackID: id}.Apply(ctx)
	case game.ZoneBattlefield:
		return BounceToHand{Target: id}.Apply(ctx)
	}
	return nil
}
