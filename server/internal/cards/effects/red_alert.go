package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Alert! — Instant {2}{R} (Star Trek, #4973):
//
//	"This spell costs {2} less to cast if a creature is attacking you.
//	 Creatures you control get +2/+0 until end of turn."
//
// The discount (CR 601.2f) is Heroic Return's: a creature attacking the
// caster themselves, not a planeswalker or battle they control. The
// +2/+0 is fixed over the creatures the caster controls as it resolves
// (CR 611.2c); a creature that arrives later does not get it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4636afb5-453a-4f27-8916-1cbc32459ba6",
		Name:         "Red Alert!",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if a creature is attacking you.", aCreatureIsAttackingYou()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Red Alert! — creatures you control get +2/+0"}.Apply(ctx)
		},
	})
}
