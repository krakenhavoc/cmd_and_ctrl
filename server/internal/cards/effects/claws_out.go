package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claws Out — Instant {3}{W}{W} (Foundations, #4465):
//
//	"Affinity for Cats (This spell costs {1} less to cast for each Cat you
//	 control.)
//	 Creatures you control get +2/+2 until end of turn."
//
// Polliwallop's affinity for a creature type (it reduces only the
// generic part), then Overrun's team pump without the trample. The pump
// reaches the creatures you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "499d76c1-97ff-4006-922c-be1684a7754c",
		Name:         "Claws Out",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Cats", HasSubtype("Cat")),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match:     And(Creature(), YouControl()),
				Power:     2,
				Toughness: 2,
				Label:     "Claws Out — +2/+2",
			}.Apply(ctx)
		},
	})
}
