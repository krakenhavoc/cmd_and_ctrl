package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vitalizing Wind — Instant {8}{G} (Prophecy, #5617):
//
//	"Creatures you control get +7/+7 until end of turn."
//
// Pride of Conquerors' team pump at a fixed +7/+7. The creatures are
// locked in as the spell resolves (CR 611.2c), so one that arrives
// later gets nothing, and the boost ends in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98b802f3-4785-41af-adfe-e77d0ae4648c",
		Name:         "Vitalizing Wind",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 7, Toughness: 7, Label: "Vitalizing Wind — +7/+7"}.Apply(ctx)
		},
	})
}
