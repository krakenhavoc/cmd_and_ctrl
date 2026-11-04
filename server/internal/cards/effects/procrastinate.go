package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Procrastinate — Sorcery {X}{U} (Secrets of Strixhaven, #4456):
//
//	"Tap target creature. Put twice X stun counters on it. (If a permanent
//	 with a stun counter would become untapped, remove one from it instead.)"
//
// X is announced as the spell is cast and read back at resolution
// (Context.X); the target is tapped, then given 2X stun counters as one
// placement (so a counter replacement applies once, and a creature that was
// already tapped still gets them). A target that is illegal on resolution
// makes the whole spell do nothing (CR 608.2b). The stun counter's untap
// replacement is the engine's (CR 122.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f8fa68e-e961-4ae1-b27d-22d87e90f26b",
		Name:         "Procrastinate",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok || !onBattlefield(ctx.Game, id) {
				return nil
			}
			if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return AddCounter{Target: id, Kind: game.CounterStun, N: 2 * ctx.X()}.Apply(ctx)
		},
	})
}
