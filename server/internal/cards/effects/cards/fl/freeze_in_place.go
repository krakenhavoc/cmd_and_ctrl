package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freeze in Place — Sorcery {1}{U} (Wilds of Eldraine, #4471):
//
//	"Tap target creature an opponent controls and put three stun counters
//	 on it. Scry 2. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// The tap and the three stun counters go on the one target; the stun
// counters' untap replacement is the engine's (CR 122.1d). Then scry 2
// (CR 701.22a). A target that has become illegal counters the whole
// spell, scry included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bf90bb0-f9c1-49c8-905e-32747e921e92",
		Name:         "Freeze in Place",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterStun, N: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
