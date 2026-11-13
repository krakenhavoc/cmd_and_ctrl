package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warrior's Honor — Instant {2}{W} (Visions, #5631):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// The creatures you control as it resolves (CR 611.2c), so one that
// enters later this turn is not pumped. A layer 7c change (CR 613.4c)
// that ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3c9da49-ef4f-48c7-99c8-0f226b485285",
		Name:         "Warrior's Honor",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Warrior's Honor — creatures you control get +1/+1",
			}.Apply(ctx)
		},
	})
}
