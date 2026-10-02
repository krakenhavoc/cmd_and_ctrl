package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Sculpt — Instant {1}{U}{U} (EDHREC rank 3960):
//
//	"Counter target spell. If you control a Wizard, add an amount of
//	 {C} equal to the amount of mana spent to cast that spell at the
//	 beginning of your next main phase."
//
// Mana Drain with a tribal tax: in a Wizard deck it is the best
// counterspell in the format, and everywhere else it is Counterspell
// for an extra mana. It is written as Mana Drain is (mana_drain.go) —
// the same "your next main phase" step pick, the same delayed refund
// body — with two differences, both printed.
//
// "The amount of MANA SPENT to cast that spell" is not the spell's
// mana value. The two come apart on every spell that was taxed, made
// cheaper, cast for an alternative cost, or cast with {X} — a Thalia
// tax adds to the mana spent and not to the mana value, a convoked
// spell's tapped creatures spend no mana at all, and a spell cast
// without paying its mana cost spent none (CR 601.2h). It is read off
// the countered spell's own payment record (StackItem.Paid, #761)
// while the spell is still on the stack and BEFORE CounterTarget runs
// — the last moment it is findable — through the same
// ManaSpent.Total() view every mana-spent reader uses. A copy of a
// spell was not cast and spent nothing (CR 707.10), so it refunds
// nothing.
//
// "If you control a Wizard" is checked as Mana Sculpt resolves, which
// is when the delayed trigger is created or not. ControlsA reads
// effective subtypes, so a changeling counts. As with Mana Drain, a
// spell that can't be countered still pays out.
//
// This card used to ship without the refund at all: its comment said
// the engine kept no record of the mana spent on a spell. #761 built
// exactly that record; #1735 found the stale caveat.
//
// Declared caveat, the one every mana-spent reader in the catalog
// carries: a spell its caster cast with strict mana off was not
// charged by the engine (PaidCost.OnPaper), so the amount is unknown
// and reads as zero — no refund, the weaker-than-printed answer
// ADR 0068 §3 requires.
func init() {
	Register(Spec{
		OracleID:     "35e2f82e-7ca3-4a92-9134-b7999eef5337",
		Name:         "Mana Sculpt",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If the countered spell was cast with strict mana off, the game doesn't know how much mana was spent on it, so Mana Sculpt adds no mana for it.",
		},
		Targets: TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			spent := ctx.Game.StackItemPaidForEffect(stackID).Spent().Total()
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if spent <= 0 || !ControlsA("Wizard")(ctx.Game, item.Controller) {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:                 manaDrainNextMainPhaseStep(ctx.Game, item.Controller),
				ControllerTurnOnly: true,
				Label:              "Mana Sculpt — add {C} × " + strconv.Itoa(spent),
				Body:               manaDrainRefundBody,
				Params:             game.EffectParams{Amount: spent},
			}.Apply(ctx)
		},
	})
}
