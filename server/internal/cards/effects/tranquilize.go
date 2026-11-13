package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tranquilize — Sorcery {1}{U} (Assassin's Creed, #5466):
//
//	"Tap target creature an opponent controls and put three stun
//	 counters on it. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// Freeze in Place without the scry. The stun counter's replacement of
// the untap is the engine's (CR 122.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1133633a-b717-42cb-b477-4e18584d876f",
		Name:         "Tranquilize",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 3}.Apply(ctx)
		},
	})
}
