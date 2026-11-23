package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tyranid Invasion — Sorcery {3}{G} (Warhammer 40,000 Commander, #5948):
//
//	"Create a number of 3/3 green Tyranid Warrior creature tokens with
//	 trample equal to the number of opponents you have."
//
// An opponent who has left the game is not an opponent (CR 800.4a), so the
// count is the opponents still in it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33a45c4a-1481-4c7e-b823-435cb84231e8",
		Name:         "Tyranid Invasion",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 0
			effects.EachOpponent(ctx.Game, ctx.Controller(), func(uuid.UUID) bool { n++; return false })
			if n == 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("3/3 green Tyranid Warrior with trample"), N: n}.Apply(ctx)
		},
	})
}
