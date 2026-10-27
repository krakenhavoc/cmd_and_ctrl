package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_to_hand.go — the continuation shared by every "choose a
// card in a graveyard; it goes to its owner's hand" clause.
//
// Its own file rather than a tail on helpers.go, per the convention
// enters_tapped.go set: concurrent card batches collide on shared
// files.

// returnPickedToHand is the keyed continuation (ChooseCardsPrompt.ThenKey)
// of a pick over a graveyard whose answer goes to its owner's hand —
// Skullwinder's opponent's mandatory half, Ripples of Undeath's "put a
// card from among those cards into your hand", Takenuma's channel. The
// prompt carries the asking item (ChooseCardsPrompt.Asked), so the body
// runs against a Context rebuilt from it and a table waiting on the pick
// is a restore point (#4253). An on-disk key: never renamed, never reused.
//
// An empty pick is a legal answer to any prompt with a floor of zero
// and moves nothing.
var returnPickedToHand = askedPick("choose-cards/return-picked-to-hand", func(ctx *Context, r game.CardsChosen) error {
	return ReturnPickedToHandNow(ctx, r.Picked)
})

// queueReturnPickedToHand asks `chooser` for one of
// `cards`, which sit in a graveyard, and returns the picked card to its
// owner's hand: returnPickedToHand, with the graveyard re-checked on
// submit, since a card can leave it between the question and the answer.
func queueReturnPickedToHand(ctx *Context, chooser uuid.UUID, question string, cards []uuid.UUID) {
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chooser,
		FromPlayer: chooser,
		Source:     ctx.Source(),
		Question:   question,
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    returnPickedToHand,
		Asked:      game.NewAskedItem(ctx.Item),
	})
}

// ReturnPickedToHandNow returns each picked graveyard card to its owner's
// hand, for a keyed body that already holds the answer (Smuggler's
// Surprise).
func ReturnPickedToHandNow(ctx *Context, picked []uuid.UUID) error {
	for _, id := range picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
