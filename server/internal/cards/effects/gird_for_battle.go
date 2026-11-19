package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gird for Battle — Sorcery {W} (Foundations Jumpstart):
//
//	"Put a +1/+1 counter on each of up to two target creatures."
//
// Basri's Acolyte's counters on up to two targets, without the "you
// control" or "other" limits: each target still legal on resolution
// gets one counter (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc72d463-8818-4c19-bd5a-f47841d1adb0",
		Name:         "Gird for Battle",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return putPlusOneCounterOnEachLegalTarget(ctx.Game, item)
		},
	})
}
