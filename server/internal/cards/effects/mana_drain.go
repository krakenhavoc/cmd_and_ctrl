package effects

import (
	"strconv"
	"strings"

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
// Sandbox simplification: it fires at the beginning of the
// controller's next PRECOMBAT main phase. A Drain cast during your
// own precombat main is printed to pay out in that turn's
// postcombat main; here it waits for the next turn, because a
// delayed trigger names one step and there is no "whichever main
// phase comes first" shape. The mana arrives later, never sooner or
// bigger — weaker than printed, and only in that one case.
func init() {
	Register(Spec{
		OracleID: "74d3277a-38e5-4732-afed-084a56148f20",
		Name:     "Mana Drain",
		Targets:  TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			mv := 0
			if victim := ctx.Game.StackItemForEffect(stackID); victim != nil {
				if card, ok := ctx.Game.LookupCardForEffect(stackID); ok {
					mv = manaValueOnStack(card, victim)
				}
			}
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if mv <= 0 {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:                 game.StepPrecombatMain,
				ControllerTurnOnly: true,
				Label:              "Mana Drain — add {C} × " + strconv.Itoa(mv),
				Effect:             manaDrainRefund(mv),
			}.Apply(ctx)
		},
	})
}

// manaDrainRefund builds the delayed trigger's effect: add mv
// colorless mana to the controller's pool. mv is a copied int, so
// the closure captures no game state and survives Clone / undo.
func manaDrainRefund(mv int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return AddMana{Produced: strings.Repeat("{C}", mv)}.Apply(NewContext(g, item))
	}
}
