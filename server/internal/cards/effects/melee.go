package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Melee — Instant {4}{R}:
//
//	"Cast this spell only during combat on your turn before blockers are
//	 declared.
//	 You choose which creatures block this combat and how those
//	 creatures block.
//	 Whenever a creature attacks and isn't blocked this combat, untap it
//	 and remove it from combat."
//
// The timing is a CastCondition: the caster's own turn, in the
// beginning of combat or declare attackers step (CR 506.8b, 506.8c:
// "during combat" is any combat phase of the turn, and "before blockers
// are declared" is before that combat's declare blockers step begins).
//
// The second sentence is a DeclarationControl for this combat, blockers
// only (game/declaration_control.go, #4154): the caster declares every
// defending player's blocks, under every restriction and requirement.
//
// The third is a delayed triggered ability with a stated duration,
// "this combat" (CR 603.7b): a repeating trigger watching each
// defender's completed block declaration (EventBlockersDeclared, the
// CR 509.3g moment), with an each-condition naming every creature
// attacking that player, or a planeswalker or battle they defend, that
// no creature blocked, so each one triggers on its own (CR 603.2c). It
// untaps that creature and removes it from combat (CR 506.4), if it is
// still that object on the battlefield (CR 603.7c). The trigger ends with
// the combat phase (UntilEndOfCombat), so an additional combat is
// untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93fc1ff8-6ef0-42a2-b090-ba95b6f3c98e",
		Name:         "Melee",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			active := g.Turn.ActiveSeat
			if active < 0 || active >= len(g.Seats) || g.Seats[active] == nil || g.Seats[active].ID != controller {
				return false
			}
			return g.Turn.Step == game.StepBeginCombat || g.Turn.Step == game.StepDeclareAttackers
		},
		CastConditionLabel: "Cast this spell only during combat on your turn before blockers are declared.",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.BeginDeclarationControlForEffect(game.DeclarationControl{
				Declarer:   item.Controller,
				Source:     item.SourceCardID,
				SourceName: "Melee",
				Blockers:   true,
				ThisCombat: true,
			})
			d, ok := ctx.Game.UntilEndOfCombatDuration()
			if !ok {
				return nil
			}
			ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
				Controller:   item.Controller,
				SourceCardID: ctx.Source(),
				Label:        "Melee — untap the unblocked attacker and remove it from combat",
				On:           []game.EventKind{game.EventBlockersDeclared},
				Condition:    eachUnblockedAttackerCondition,
				Body:         untapAndRemoveThePayloadFromCombatBody,
				Repeats:      true,
				Duration:     &d,
			})
			return nil
		},
	})
}

var (
	// "Whenever a creature attacks and isn't blocked" as a delayed
	// trigger (Melee, #4154): each creature attacking the player whose
	// block declaration just completed, that nothing blocks (CR 509.3g).
	eachUnblockedAttackerCondition = game.DelayedEachCondition("combat/each-attacker-not-blocked", eachUnblockedAttacker)

	// "Untap it and remove it from combat" (Melee, #4154): the payload's
	// creature, if it is still that object on the battlefield.
	untapAndRemoveThePayloadFromCombatBody = game.SimpleDelayedBody("combat/untap-and-remove-the-payload", untapAndRemoveThePayloadFromCombat)
)

// eachUnblockedAttacker is every creature attacking ev.Actor (the
// defender whose declaration completed), or a permanent they defend,
// that is unblocked: attacksAndIsNotBlocked's question, asked of every
// attacker.
func eachUnblockedAttacker(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventBlockersDeclared {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.AttackingTarget == uuid.Nil || !c.IsCreature() {
			continue
		}
		if g.DefendingPlayerForAttackerForEffect(c.InstanceID) != ev.Actor || !g.UnblockedAttackerForEffect(c.InstanceID) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// untapAndRemoveThePayloadFromCombat untaps the payload's creature and
// removes it from combat (CR 506.4).
func untapAndRemoveThePayloadFromCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, obj.ID)
	if !ok || c.ObjectEpoch != obj.Epoch {
		return nil
	}
	if err := (UntapTarget{Target: obj.ID}).Apply(ctx); err != nil {
		return err
	}
	if c.AttackingTarget == uuid.Nil && c.BlockingTarget == uuid.Nil {
		return nil
	}
	return g.RemoveFromCombatForEffect(obj.ID)
}
