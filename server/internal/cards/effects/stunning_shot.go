package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stunning Shot — Sorcery {1}{W} (Marvel Super Heroes Commander, #4455):
//
//	"Put two +1/+1 counters on up to one target creature you control. Tap
//	 up to one target creature an opponent controls and put a stun counter
//	 on it."
//
// Two "up to one" clauses (Clauses, read by slot: ClauseTarget), so
// choosing only the second still stuns and choosing only the first still
// grows. Each half is skipped when its target has left or stopped being
// legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c0bb9b0-e7e0-465e-826b-2e56032aed1b",
		Name:         "Stunning Shot",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
			TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && onBattlefield(ctx.Game, t.ID) {
				if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && onBattlefield(ctx.Game, t.ID) {
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
