package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodcurdling Scream — Sorcery {X}{B} (Portal Second Age, #5625):
//
//	"Target creature gets +X/+0 until end of turn."
//
// Enrage's body at sorcery speed: X is announced with the cast
// (CR 601.2b) and read as the spell resolves, and the boost ends in the
// cleanup step (CR 514.2). A target that has left by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ee2060a-5152-4a53-8183-cbd642e4cc29",
		Name:         "Bloodcurdling Scream",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(PowerXUntilEOT(0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Bloodcurdling Scream — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
}
