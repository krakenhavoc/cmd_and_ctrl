package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ambush — Instant {3}{R} (Homelands, S82 OLD b827, #5635):
//
//	"Blocking creatures gain first strike until end of turn."
//
// Fangren Pathcutter's group grant, over the blocking creatures instead of the
// attacking ones: the set is fixed as the spell resolves (CR 611.2c), so a
// creature that blocks later does not get it. Outside the declare-blockers
// window no creature is blocking and the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05828c1e-9185-4b70-94fe-753eacdf45a1",
		Name:         "Ambush",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.BlockingCreature(),
				Keywords: []string{"first strike"},
				Label:    "Ambush — first strike until end of turn",
			}.Apply(ctx)
		},
	})
}
