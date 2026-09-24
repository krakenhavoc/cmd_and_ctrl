package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Drain — Instant {U}{U} (EDHREC rank 117):
//
//	"Counter target spell. At the beginning of your next main phase,
//	 add an amount of {C} equal to that spell's mana value."
//
// Counterspell with a refund. The mana value is read off the
// countered spell WHILE IT IS STILL ON THE STACK — CR 202.3e, so an
// X spell counts its chosen X — and BEFORE CounterTarget runs, which
// is the last moment the card is findable there. The refund is a
// CR 603.7 delayed trigger that resolves through the new AddMana
// primitive; nothing is added for a mana value of 0 (a countered
// ability, a land-like spell) or for a countered ability item, which
// has no card to read.
//
// The trigger sets ControllerTurnOnly: "YOUR next main phase" means
// the next main phase on the caster's own turn, not the very next
// main phase at the table — that is the field's whole reason to
// exist (Arcane Denial's "the next turn's upkeep" is the opposite
// case and leaves it unset).
//
// "Your next main phase" is not always precombat: manaDrainNextMainPhaseStep
// reads the CURRENT step at resolution and picks whichever main phase
// is still ahead of it on the controller's own turn — this turn's
// postcombat main when the Drain resolves at or after precombat main
// but before postcombat, and precombat main otherwise (before combat
// this turn, or on any turn but the controller's own, where
// ControllerTurnOnly then waits out however many turns pass). A
// delayed trigger only fires on a step's ENTRY (see delayed.go), so
// scheduling AT the step already in progress would wait a full turn
// rather than firing later this same turn — which is exactly the case
// this picks around.
func init() {
	Register(Spec{
		OracleID:     "74d3277a-38e5-4732-afed-084a56148f20",
		Name:         "Mana Drain",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			mv := 0
			if ctx.Game.StackItemForEffect(stackID) != nil {
				if card, ok := ctx.Game.LookupCardForEffect(stackID); ok {
					mv, _ = ctx.Game.ManaValueForEffect(card)
				}
			}
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if mv <= 0 {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:                 manaDrainNextMainPhaseStep(ctx.Game, item.Controller),
				ControllerTurnOnly: true,
				Label:              "Mana Drain — add {C} × " + strconv.Itoa(mv),
				Body:               manaDrainRefundBody,
				Params:             game.EffectParams{Amount: mv},
			}.Apply(ctx)
		},
	})
}

// The refund itself is manaDrainRefundBody (delayed_bodies.go): the
// mana value it adds rides as the trigger's params (ADR 0041 phase 3,
// #1497), not as a captured int.

// manaDrainNextMainPhaseStep is "your next main phase" (CR 601 has no
// defined term for this, but the reading every judge gives it is
// "whichever main phase — precombat or postcombat — hasn't happened
// yet on your current or next turn"). Reads g.Turn and g.Seats
// directly rather than through g.ActivePlayer, which takes the read
// lock this runs under already (IsYourTurn's own contract).
func manaDrainNextMainPhaseStep(g *game.Game, controller uuid.UUID) game.Step {
	if !IsYourTurn(g, controller) {
		// Not my turn at all: the earliest main phase I can have is
		// my next precombat main, however many turns away that is —
		// ControllerTurnOnly does the waiting.
		return game.StepPrecombatMain
	}
	seq := game.TurnSequence()
	indexOf := func(target game.Step) int {
		for i, s := range seq {
			if s == target {
				return i
			}
		}
		return -1
	}
	cur, pre, post := indexOf(g.Turn.Step), indexOf(game.StepPrecombatMain), indexOf(game.StepPostcombatMain)
	switch {
	case cur < pre:
		// Untap / upkeep / draw: this turn's precombat main hasn't
		// begun yet.
		return game.StepPrecombatMain
	case cur < post:
		// At or after precombat main, but before postcombat main
		// (including precombat main itself, whose entry hook has
		// already run): this turn's postcombat main is next.
		return game.StepPostcombatMain
	default:
		// At or after postcombat main: wait for next turn's precombat
		// main.
		return game.StepPrecombatMain
	}
}
