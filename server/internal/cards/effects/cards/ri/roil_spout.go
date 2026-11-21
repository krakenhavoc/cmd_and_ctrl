package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	t := effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "ebc1d343-26f1-4d85-ac3a-610da24c990e",
		Name:         "Roil Spout",
		Completeness: effects.CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(4, "{4}{W}{U}", t),
		},
		OnResolve: effects.AwakenAfter(4, func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.AwakenSpellCardTarget(ctx); ok {
				return ctx.Game.TuckToLibraryForEffect(id, false)
			}
			return nil
		}),
	})
}
