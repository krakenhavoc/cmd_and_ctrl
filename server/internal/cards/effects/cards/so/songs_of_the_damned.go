package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Songs of the Damned — Instant {B} (Ice Age, S82 OLD batch 851, #5636):
//
//	"Add {B} for each creature card in your graveyard."
//
// Battle Hymn's instant that uses the stack, counting creature cards in your
// graveyard as it resolves (Grave Strength's count): a player with none adds
// nothing, and the Songs themselves are not in the graveyard yet.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39419d29-47e7-4d71-953b-96a2e14ae964",
		Name:         "Songs of the Damned",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
			if n == 0 {
				return nil
			}
			return effects.AddMana{Amount: effects.ManaN(n, "B")}.Apply(ctx)
		},
	})
}
