package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceaseless Conflict — Sorcery {3}{W}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Destroy all creatures. Then create a 3/2 red and white Spirit
//	 creature token for each nontoken creature you controlled that was
//	 destroyed this way."
//
// Kaya's Wrath's shape. DestroyAllMatching hands its tail the pre-move
// copies of the creatures that were really destroyed, so a regenerated
// or indestructible creature is not counted, and the count reads each
// copy's controller and token status as it last existed on the
// battlefield (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f109125-b743-4ecf-84da-690d8e8ddf9c",
		Name:         "Ceaseless Conflict",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Creature(), ThenTail: ctx.ItemTail(ceaselessConflictSpirits)}.Apply(ctx)
		},
	})
}

var ceaselessConflictSpirits = effects.MassTail("mass/ceaseless-conflict-spirits", func(ctx *effects.Context, swept []game.Card, _ int) error {
	n := 0
	for _, c := range swept {
		if c.Controller == ctx.Controller() && !effects.IsToken(c) {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("3/2 red and white Spirit"), N: n}.Apply(ctx)
})
