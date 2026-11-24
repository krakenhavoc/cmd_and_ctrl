package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ReturnExiledCreaturesTappedBody is a flicker's delayed return for "return it
// to the battlefield tapped under its owner's control": the listed cards come
// back together under their owners' control, and a card that is a creature as
// it enters comes in tapped (CR 614.12), judged card by card (Nezahal, Primal
// Tide). An on-disk key: never renamed, never reused.
var ReturnExiledCreaturesTappedBody = game.SimpleDelayedBody(
	"flicker/return-exiled-creatures-tapped", returnExiledCardsCreaturesTapped)

func returnExiledCardsCreaturesTapped(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return ReturnFromExileTogether{Targets: ids, TappedIfType: []string{"creature"}}.Apply(NewContext(g, item))
}

// ExileThisThenReturnItTappedAtTheNextEndStep is "Exile <this>. Return it to
// the battlefield tapped under its owner's control at the beginning of the next
// end step": the source leaves, and a delayed trigger (CR 603.7) brings the
// card that reached exile back. A source that has already left, or is a new
// object, is not exiled.
func ExileThisThenReturnItTappedAtTheNextEndStep(ctx *Context, label string) error {
	id := ctx.Source()
	if id == uuid.Nil || ctx.IsNewSourceObject(id) || !OnBattlefield(ctx.Game, id) {
		return nil
	}
	return exileCardsThenScheduleReturnBody(ctx, []uuid.UUID{id}, label, ReturnExiledCreaturesTappedBody)
}
