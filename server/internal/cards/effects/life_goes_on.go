package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Life Goes On — Instant {G} (Hour of Devastation, #5534):
//
//	"You gain 4 life. If a creature died this turn, you gain 8 life
//	 instead."
//
// The condition is read as the spell resolves, from the turn's death count
// (Gruesome Discovery's b11CreaturesDiedThisTurn), any player's creature,
// token or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15d2176b-ab3d-4738-bcbc-bcebf891d16e",
		Name:         "Life Goes On",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 4
			if b11CreaturesDiedThisTurn(ctx.Game) > 0 {
				n = 8
			}
			return GainLife{Amount: n}.Apply(ctx)
		},
	})
}
