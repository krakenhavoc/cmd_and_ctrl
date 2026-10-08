package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roil Spout — Sorcery {1}{W}{U}:
//
//	"Put target creature on top of its owner's library.
//	 Awaken 4—{4}{W}{U}"
//
// ADR 0135 §3 (#2411): the tuck, then the awaken land (CR 702.113a), each
// target checked on its own (CR 608.2b).
//
// No simplifications.
func init() {
	t := TargetCreature("target creature")
	Register(Spec{
		OracleID:     "ebc1d343-26f1-4d85-ac3a-610da24c990e",
		Name:         "Roil Spout",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(4, "{4}{W}{U}", t),
		},
		OnResolve: AwakenAfter(4, func(_ *game.StackItem, ctx *Context) error {
			if id, ok := awakenSpellCardTarget(ctx); ok {
				return ctx.Game.TuckToLibraryForEffect(id, false)
			}
			return nil
		}),
	})
}
