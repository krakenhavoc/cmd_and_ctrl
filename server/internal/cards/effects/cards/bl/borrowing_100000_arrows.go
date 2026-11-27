package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borrowing 100,000 Arrows — Sorcery {2}{U} (Commander 2013, S82 OLD b380,
// #5560):
//
//	"Draw a card for each tapped creature target opponent controls."
//
// The count is taken as the spell resolves, from the tapped creatures the
// target opponent controls (CR 608.2b: a target that became illegal counters
// the whole spell).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3f10772-d055-419d-a535-4f094eaf437c",
		Name:         "Borrowing 100,000 Arrows",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: effects.S82OldB380TappedCreaturesOf(ctx.Game, t.ID)}.Apply(ctx)
		},
	})
}
