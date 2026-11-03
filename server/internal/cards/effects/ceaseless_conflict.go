package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2f109125-b743-4ecf-84da-690d8e8ddf9c",
		Name:         "Ceaseless Conflict",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Creature(), ThenTail: ctx.ItemTail(ceaselessConflictSpirits)}.Apply(ctx)
		},
	})
}

var ceaselessConflictSpirits = MassTail("mass/ceaseless-conflict-spirits", func(ctx *Context, swept []game.Card, _ int) error {
	n := 0
	for _, c := range swept {
		if c.Controller == ctx.Controller() && !IsToken(c) {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/2 red and white Spirit"), N: n}.Apply(ctx)
})
