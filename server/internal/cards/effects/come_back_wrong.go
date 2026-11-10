package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Come Back Wrong — Sorcery {2}{B} (Duskmourn: House of Horror, #4466):
//
//	"Destroy target creature. If a creature card is put into a graveyard
//	 this way, return it to the battlefield under your control. Sacrifice
//	 it at the beginning of your next end step."
//
// The destruction is the keyed batch form, whose tail is told what really
// landed in a graveyard: a creature saved by indestructible, regeneration
// or a replacement that sent it elsewhere is not returned, and a token is
// not a card. The return and the delayed sacrifice happen inside the
// resolution, so nobody can take the card out of the graveyard in between
// (the card's ruling). The sacrifice is "your next end step", the
// controller's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a9783c5-084b-4cd3-b7fa-ae9d7fed51d1",
		Name:         "Come Back Wrong",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{item.Targets[0].ID}, ctx.ItemTail(stdB26ComeBackWrongDestroyed))
		},
	})
}
