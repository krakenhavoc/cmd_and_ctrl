package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impede Momentum — Sorcery {1}{U} (Dominaria United, #5487):
//
//	"Tap target creature and put three stun counters on it. Scry 1. (If a
//	 permanent with a stun counter would become untapped, remove one from
//	 it instead.)"
//
// Any creature may be the target. The tap and the three stun counters go on
// it; the counters' untap replacement is the engine's (CR 122.1d). Then
// scry 1 (CR 701.22a). A target that has become illegal counters the whole
// spell, scry included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0239ff1-6ae6-431a-806e-7321d8272a1d",
		Name:         "Impede Momentum",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
