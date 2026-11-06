package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8bf90bb0-f9c1-49c8-905e-32747e921e92",
		Name:         "Freeze in Place",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Targets:      TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
