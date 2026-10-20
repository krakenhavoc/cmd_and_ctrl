package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surprise Deployment — Instant {3}{W}:
//
//	"Cast this spell only during combat.
//	 You may put a nonwhite creature card from your hand onto the
//	 battlefield. At the beginning of the next end step, return that
//	 creature to your hand. (Return it only if it's on the battlefield.)"
//
// Three pieces, each an existing shape:
//
//   - "Cast this spell only during combat" is a CastCondition on the
//     combat phase (every step from beginning of combat through end of
//     combat), the Take the Bait / Blaze of Glory gate. It is a gate on
//     CASTING, so the spell is not offered outside combat and a cast
//     attempt is refused; once cast, nothing re-checks it at resolution.
//   - The put is PutFromHandOntoBattlefield with Optional ("you MAY"):
//     it is not a cast, so no mana is paid and no cast trigger fires,
//     but the creature really enters (CR 614 entry pipeline, its ETB
//     triggers fire). A nonwhite creature is any creature card that is
//     not white — a colorless one qualifies. A permanent card that is
//     not a creature, or a white creature, is never offered.
//   - The return is a CR 603.7 delayed trigger at the beginning of the
//     NEXT end step (any player's: the clause says "the next end step",
//     not "your"), scheduled only when a creature actually entered. It
//     reuses Zara, Renegade Recruiter's body, which bounces the creature
//     only if it is still the very object that entered (CR 400.7): a
//     creature that was flickered, died or was already bounced is a
//     different object or not on the battlefield, and the parenthetical
//     says "return it only if it's on the battlefield". The creature
//     goes to its owner's hand — the caster's, since it came from there.
//
// The return trigger goes on the stack at the end step like any
// trigger, so opponents can respond, and the creature's own ETB, attack
// and combat-damage triggers are all live in the meantime. That is the
// point of the card: an ETB engine in combat, or a surprise blocker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61a1748c-e772-4338-9edd-720cd5d2fc12",
		Name:         "Surprise Deployment",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
		},
		CastConditionLabel: "Cast this spell only during combat.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutFromHandOntoBattlefield{
				Match:    And(Creature(), Not(OfColor("W"))),
				Optional: true,
				Label:    "Surprise Deployment — you may put a nonwhite creature card from your hand onto the battlefield",
				Then:     surpriseDeploymentReturn,
			}.Apply(ctx)
		},
	})
}

// surpriseDeploymentReturn is the Then of the put: with a creature on
// the battlefield, schedule its return to hand at the next end step.
// Package-level, so the continuation captures nothing.
func surpriseDeploymentReturn(g *game.Game, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(res.Entered)
	if !ok {
		return nil
	}
	// The same record ScheduleDelayedTrigger.Apply writes (a Then has a
	// Game and the result, not a Context): At defaults to the next end
	// step, whoever's it is.
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   res.Player,
		SourceCardID: res.Source,
		Label:        "Surprise Deployment — return that creature to your hand",
		At:           game.StepEnd,
		Cards:        []uuid.UUID{res.Entered},
		Body:         zaraReturnToHandBody,
		Params:       game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	})
	return nil
}
