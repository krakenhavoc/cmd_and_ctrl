package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blur — Instant {2}{U} (Commander Legends: Battle for Baldur's Gate,
// #5490):
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control.
//	 Draw a card."
//
// Personify's shape with a draw in place of the token: Ephemerate's
// flicker (flickerTheTargetToItsOwner), so the card comes back under its
// OWNER's control as a new object (CR 400.7) and a token exiled this way
// does not return (CR 111.8). Then you draw. If the target is illegal as
// Blur resolves, the spell does nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00b36996-43c6-42a5-892c-c7c8864cf973",
		Name:         "Blur",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.FlickerTheTargetToItsOwner(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
