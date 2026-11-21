package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "15d2176b-ab3d-4738-bcbc-bcebf891d16e",
		Name:         "Life Goes On",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 4
			if effects.B11CreaturesDiedThisTurn(ctx.Game) > 0 {
				n = 8
			}
			return effects.GainLife{Amount: n}.Apply(ctx)
		},
	})
}
