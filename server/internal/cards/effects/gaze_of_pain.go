package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gaze of Pain — Sorcery {1}{B}:
//
//	"Until end of turn, whenever a creature you control attacks and isn't
//	 blocked, you may choose to have it deal damage equal to its power to
//	 a target creature. If you do, it assigns no combat damage this turn."
//
// #3531, ADR 0187 amendment 2026-10-16: a repeating delayed triggered
// ability (CR 603.7a-b) that triggers once for each creature you control
// that attacks and isn't blocked (CR 509.3, 603.2), with that creature as
// its "it" (an each-condition, game.DelayedEachCondition). The target is
// chosen as each instance goes on the stack (CR 603.3d), from the body's
// own registration (game.TargetedDelayedBody). "You may" is asked as it
// resolves; the creature deals the damage with its power then, or its
// last-known power if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3a479d4-3e36-496b-abf8-b83231628d76",
		Name:         "Gaze of Pain",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DelayedOnEvent{
				Label:     "Gaze of Pain — you may have it deal damage equal to its power to target creature",
				On:        []game.EventKind{game.EventBlockersDeclared},
				Condition: eachCreatureYouControlAttacksUnblockedCondition,
				Body:      gazeOfPainBody,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}

// gazeOfPain is the fired trigger's body: "it" is the payload, the target
// the clause the instance went on the stack with.
func gazeOfPain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	it := firstPayloadCard(ctx)
	target := firstLegalCardTarget(ctx)
	if it == uuid.Nil || target == uuid.Nil {
		return nil
	}
	return MayChoice{
		Question: "Gaze of Pain — have the attacking creature deal damage equal to its power to the target creature? (if you do, it assigns no combat damage this turn)",
		Then:     gazeOfPainAnswer,
		Carry:    []uuid.UUID{it, target},
	}.Apply(ctx)
}

// gazeOfPainAnswer is the "yes": the attacking creature (the first carried
// ID) deals damage equal to its power to the target (the second), then
// assigns no combat damage this turn. Keyed (#4291).
var gazeOfPainAnswer = MayYes("confirm/gaze-of-pain-damage", func(ctx *Context, r game.ConfirmAnswer) error {
	it, target := carryAt(r, 0), carryAt(r, 1)
	if err := (DealDamage{Source: it, Target: target, Amount: powerNowOrLastKnown(ctx.Game, it)}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), it, "Gaze of Pain — assigns no combat damage this turn")
	return nil
})
