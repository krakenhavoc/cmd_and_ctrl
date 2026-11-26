package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revealing Wind — Instant {2}{G} (Dragons of Tarkir):
//
//	"Prevent all combat damage that would be dealt this turn. You may look at
//	 each face-down creature that's attacking or blocking."
//
// The look is a one-shot read at resolution (CR 708.5) of the face-down creatures
// in combat then, with no standing grant: a creature turned face down or entering
// the fight afterwards is not seen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "905a6d6c-5ef8-4afa-89bb-9d85a854743f",
		Name:         "Revealing Wind",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.PreventAllCombatDamageThisTurn{Label: "Revealing Wind: prevent all combat damage"}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.LookAtFaceDownCombatantsForEffect(ctx.Controller())
			return nil
		},
	})
}
