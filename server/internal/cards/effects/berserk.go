package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Berserk — Instant {G}:
//
//	"Cast this spell only before the combat damage step.
//	 Target creature gains trample and gets +X/+0 until end of turn,
//	 where X is its power. At the beginning of the next end step,
//	 destroy that creature if it attacked this turn."
//
// THE TIMING. A CastCondition checked once, at announce (CR 601.2):
// the spell may be cast in the beginning phase, the precombat main
// phase, the beginning of combat step, the declare attackers step and
// the declare blockers step, and never once a combat damage step has
// begun this turn — the first-strike damage step counts, since it is
// a combat damage step (CR 510.4). So it can't be cast in a combat
// damage step, at end of combat, in the postcombat main phase or in
// the ending phase. In a turn with more than one combat (an extra
// combat phase, CR 500.8) the window is per combat: the beginning of
// combat, declare attackers and declare blockers steps of each combat
// are before THAT combat's damage step, so the spell can be cast there
// even though an earlier combat dealt damage.
//
// THE PUMP. X is the target's power as Berserk resolves, read once
// and unclamped (a negative power gives -X/+0); trample and +X/+0 are
// one effect at one timestamp, pinned to that object (CR 400.7).
//
// THE DELAYED TRIGGER. Scheduled only if the spell resolves (CR
// 603.7a), for the next end step, naming the creature as the object it
// was (ObjectRef), so a creature that left and came back is a new
// object and is not destroyed. "If it attacked this turn" is not an
// intervening if (it doesn't follow the trigger condition, CR 603.4):
// the trigger always goes on the stack and checks as it resolves,
// against the per-object attack history. Whether it is still a
// creature doesn't matter (the 2026-03-20 ruling): an attacker that
// stopped being a creature is still destroyed. Regeneration (CR
// 701.8c) and indestructible apply as usual.
func init() {
	Register(Spec{
		OracleID:     "8b67d192-9a05-4a47-82ae-5fc4b7834d88",
		Name:         "Berserk",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return berserkBeforeCombatDamage(g)
		},
		CastConditionLabel: "Cast this spell only before the combat damage step.",
		Targets:            TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if err := untilEndOfTurn(ctx, id, nil, "Berserk — trample and +X/+0",
				game.ModifyPTMod(c.PowerForComparison(), 0), game.AddKeywordsMod("trample")); err != nil {
				return err
			}
			return ScheduleDelayedTrigger{
				Label:  "Berserk — destroy that creature if it attacked this turn",
				Cards:  []uuid.UUID{id},
				Body:   berserkDestroyIfAttackedBody,
				Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			}.Apply(ctx)
		},
	})
}

// berserkBeforeCombatDamage is the cast window: a step before combat
// damage. A combat step before damage is before its own combat's
// damage step, whatever an earlier combat did; the beginning phase and
// the precombat main phase need no earlier damage step this turn.
func berserkBeforeCombatDamage(g *game.Game) bool {
	switch g.Turn.Step {
	case game.StepBeginCombat, game.StepDeclareAttackers, game.StepDeclareBlockers:
		return true
	case game.StepUpkeep, game.StepDraw, game.StepPrecombatMain:
	default:
		return false
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventStepBegan &&
			(ev.Step == game.StepFirstStrikeDamage || ev.Step == game.StepCombatDamage) {
			return false
		}
	}
	return true
}

// berserkDestroyIfAttacked is the delayed trigger's body
// (berserkDestroyIfAttackedBody): destroy the creature p.Object names
// if it is still that object on the battlefield and it attacked this
// turn.
func berserkDestroyIfAttacked(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, id) {
		return nil
	}
	if g.TimesAttackedThisTurn(id) == 0 {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(NewContext(g, item))
}
