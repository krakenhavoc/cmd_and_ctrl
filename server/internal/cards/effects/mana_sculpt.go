package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Sculpt — Instant {1}{U}{U} (EDHREC rank 3960):
//
//	"Counter target spell. If you control a Wizard, add an amount of
//	 {C} equal to the amount of mana spent to cast that spell at the
//	 beginning of your next main phase."
//
// Mana Drain with a tribal tax: in a Wizard deck it is the best
// counterspell in the format, and everywhere else it is Counterspell
// for an extra mana.
//
// # Declared simplification (weaker than printed): no refund
//
// "The amount of MANA SPENT to cast that spell" is not the spell's
// mana value. The two come apart on every spell that was taxed, made
// cheaper, cast for an alternative cost, or cast with {X} — a Thalia
// tax adds to the mana spent and not to the mana value, a convoked
// spell's tapped creatures spend no mana at all, and Mana Drain's own
// oracle text was changed to say "mana value" precisely because
// "mana spent" needs a record nothing keeps.
//
// The engine does not keep that record (#761): mana is deducted from
// a pool at cast time and nothing attributes the deduction to a stack
// item afterwards. Guessing with mana value would be a different
// number, and in the direction that matters — a taxed spell would
// refund LESS, but a cost-reduced or convoked one would refund MORE
// than paper does, which is the direction #259 forbids.
//
// So the refund is dropped entirely and Mana Sculpt ships as a
// three-mana hard counter. Nothing about the counter is simplified:
// it is a real counter, the spell goes to its owner's graveyard, and
// a "can't be countered" spell is still a legal target that the
// counter simply does nothing to (CR 701.6a).
//
// The Wizard condition is not modelled either, because it gates only
// the clause that is gone.
func init() {
	Register(Spec{
		OracleID:     "35e2f82e-7ca3-4a92-9134-b7999eef5337",
		Name:         "Mana Sculpt",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The colorless mana refund on your next main phase isn't implemented — Mana Sculpt is a straight counterspell.",
		},
		Targets: TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
