package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// No Witnesses — Sorcery {2}{W}{W} (Murders at Karlov Manor, #4469):
//
//	"Each player who controls the most creatures investigates. Then
//	 destroy all creatures. (To investigate, create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")"
//
// The creatures are counted as the spell resolves, before anything is
// destroyed. Every player tied for the most investigates, so with no
// creatures on the battlefield every player in the game is tied at zero
// and each gets a Clue. The sweep is Day of Judgment's
// (wrathDestroyAllCreatures: simultaneous, regeneration allowed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0639bd82-084f-47d5-b158-05dfc29efa40",
		Name:         "No Witnesses",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			creatures := map[uuid.UUID]int{}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					creatures[c.Controller]++
				}
			}
			most := -1
			perSeat := make([]int, len(ctx.Game.Seats))
			for i, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				perSeat[i] = creatures[p.ID]
				if perSeat[i] > most {
					most = perSeat[i]
				}
			}
			for i, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated || perSeat[i] != most {
					continue
				}
				if err := (Investigate{Player: p.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return wrathDestroyAllCreatures(item, ctx)
		},
	})
}
