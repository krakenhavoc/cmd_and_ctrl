package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Polymorph — Sorcery {3}{U} (Magic 2010, #5585):
//
//	"Destroy target creature. It can't be regenerated. Its controller reveals
//	 cards from the top of their library until they reveal a creature card.
//	 The player puts that card onto the battlefield, then shuffles all other
//	 cards revealed this way into their library."
//
// Transmogrify's reveal with a destroy in front. The controller is read as the
// spell resolves, before the creature dies, and the reveal happens whether or
// not the creature was destroyed (an indestructible one still sends its
// controller digging).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "039115f2-6322-4674-bc64-f21883ed375a",
		Name:         "Polymorph",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			id := legal[0].ID
			player, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return effects.OldB471RevealUntilPutOntoBattlefield(ctx, player, func(c game.Card) bool { return c.IsCreature() },
				"Polymorph — reveal until a creature card")
		},
	})
}
