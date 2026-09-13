package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fell the Mighty — Sorcery {4}{W} (EDHREC rank 1472):
//
//	"Destroy all creatures with power greater than target creature's
//	 power."
//
// The small-creature deck's wrath: point it at your 1/1 and every
// bigger creature dies, the target included only if its own power is
// somehow greater than itself, which it never is. The target's power
// is read at resolution — current power, counters and anthems in —
// and the sweep is one simultaneous destruction of every creature
// above it, the caster's own included, as printed. If the target has
// left, the spell fizzles like any targeted spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1fbb2a7c-8093-4729-b8fe-cf032d99470f",
		Name:         "Fell the Mighty",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			target, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID)
			if !ok {
				return nil
			}
			return DestroyAllMatching{Match: b13PowerGreaterThan(target.CurrentPower())}.Apply(ctx)
		},
	})
}
