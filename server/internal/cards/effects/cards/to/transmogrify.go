package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transmogrify — Sorcery {3}{R} (Core Set 2021, #5511):
//
//	"Exile target creature. That creature's controller reveals cards from the
//	 top of their library until they reveal a creature card. That player puts
//	 that card onto the battlefield, then shuffles the rest into their
//	 library."
//
// Chaos Mutation's shape for one target, read as Proteus Staff does: the
// controller is read before the creature is exiled, and the reveal runs
// whether or not the exile happened. The rest never left the library, so
// shuffling it is a shuffle of the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c1c3a0a-5782-4b7d-8b9c-61eea913166d",
		Name:         "Transmogrify",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			id := legal[0].ID
			player, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			run, hit := effects.RevealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
				"Transmogrify — reveal until a creature card")
			return transmogrifyPutCreature(ctx, player, run, hit)
		},
	})
}

func transmogrifyPutCreature(ctx *effects.Context, player uuid.UUID, run []uuid.UUID, hit uuid.UUID) error {
	return ctx.Game.PutByEnteringPlayerForEffect(func() error {
		return effects.PutFromLibraryOntoBattlefield{
			Player:   player,
			Cards:    run,
			Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
			All:      true,
			ThenTail: effects.S82OldB109ShuffleRestTail,
		}.Apply(ctx)
	})
}
