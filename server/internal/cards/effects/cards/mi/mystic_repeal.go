package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mystic Repeal — Instant {G} (Theros Beyond Death, #5514):
//
//	"Put target enchantment on the bottom of its owner's library."
//
// The same tuck as Bant Charm's creature mode (Game.TuckToLibraryForEffect with
// bottom set): the card goes to the bottom of its OWNER's library (CR 401.1), a
// token ceasing to exist there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f453aca-e14f-4460-9334-f0c3f4256a0e",
		Name:         "Mystic Repeal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			return ctx.Game.TuckToLibraryForEffect(id, true)
		},
	})
}
