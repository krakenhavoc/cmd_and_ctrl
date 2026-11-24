package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Echoing Return — Sorcery {B} (Modern Horizons 2, #5501):
//
//	"Return target creature card and all other cards with the same name as
//	 that card from your graveyard to your hand."
//
// The name is read off the target as the spell resolves, and every other card
// in the caster's graveyard with that name comes back with it (a card in
// another player's graveyard does not). If the target left in response the
// whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "591fe685-767e-4e6e-905e-cc8b423caa19",
		Name:         "Echoing Return",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			target, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
			if p == nil || p.Graveyard == nil {
				return nil
			}
			ids := []uuid.UUID{id}
			for _, c := range p.Graveyard.Cards {
				if c.InstanceID != id && c.Name == target.Name {
					ids = append(ids, c.InstanceID)
				}
			}
			return effects.ReturnPickedToHandNow(ctx, ids)
		},
	})
}
