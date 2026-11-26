package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shape Anew — Sorcery {3}{U} (Scars of Mirrodin, S82 OLD batch 471, #5577):
//
//	"The controller of target artifact sacrifices it, then reveals cards from
//	 the top of their library until they reveal an artifact card. That player
//	 puts that card onto the battlefield, then shuffles all other cards
//	 revealed this way into their library."
//
// Transmogrify's shape for an artifact: the controller is read before the
// sacrifice, and the reveal runs whether or not the sacrifice happened (a
// permanent that can't be sacrificed stays, and the controller still reveals).
// The artifact enters under that player's control, as they put it there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21b7cf4f-1c86-4700-be71-68c974d143b0",
		Name:         "Shape Anew",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			player, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := ctx.Game.SacrificePermanentForEffect(id); err != nil {
				return err
			}
			return effects.OldB471RevealUntilPutOntoBattlefield(ctx, player,
				func(c game.Card) bool { return c.IsArtifact() }, "Shape Anew — reveal until an artifact card")
		},
	})
}
