package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dosan's Oldest Chant — Sorcery {4}{G} (Saviors of Kamigawa, #5601):
//
//	"You gain 6 life.
//	 Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59aeeb18-d263-426f-aebc-0f687b09801b",
		Name:         "Dosan's Oldest Chant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 6, Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GainLife{Player: ctx.Controller(), Amount: 6}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
