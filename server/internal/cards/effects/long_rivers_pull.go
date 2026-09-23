package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Long River's Pull — Instant {U}{U} (EDHREC rank 1460):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Counter target creature spell. If the gift was promised, instead
//	 counter target spell."
//
// Essence Scatter that can be Counterspell for the price of a card
// to an opponent.
//
// Gift (CR 702.174, ADR 0089): the promise swaps the target clause
// (`.Instead(TargetSpell("target spell"))`, CR 702.174m), so a
// promised cast may name any spell and an unpromised one only a
// creature spell — the engine judges the targets at announce and
// again at resolution under the clause the announcement produced. The
// opponent draws before the counter (CR 702.174j); a Pull that
// fizzles because its target left gives no card, which is the same
// rule.
func init() {
	Register(Spec{
		OracleID:     "f1993767-1d07-49c8-b8dc-04ec9840a999",
		Name:         "Long River's Pull",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature spell", Creature()),
		Gift:         GiftACard().Instead(TargetSpell("target spell")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if ctx.Game.StackItemForEffect(item.Targets[0].ID) == nil {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
