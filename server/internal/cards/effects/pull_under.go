package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pull Under — Instant — Arcane {5}{B} (Champions of Kamigawa, #5603):
//
//	"Target creature gets -5/-5 until end of turn."
//
// Sadistic Slash's body; the shrink ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b99ddb27-a59a-4cd3-88b1-97c759ddbd93",
		Name:         "Pull Under",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -5, -5)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: -5, Toughness: -5, Label: "Pull Under — -5/-5"}.Apply(ctx)
		},
	})
}
