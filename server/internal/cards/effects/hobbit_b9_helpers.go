package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hobbit_b9_helpers.go — shared pieces of The Hobbit, half B, batch 9
// (#4454).

// hobB9MillFour is "mill four cards, then <the tail>": the tail runs once the
// four have landed, with the cards that really reached the graveyard.
func hobB9MillFour(ctx *Context, tail game.TailRef) error {
	return MillToZone{Player: ctx.Controller(), N: 4, ThenTail: ctx.ItemTail(tail)}.Apply(ctx)
}

// hobB9TakeFromMilled is "put <min..max> <match> card(s) from among them into
// your hand", asked of the resolving effect's controller (Speak Secrets, Silvan
// Rally). `milled` is MillToZone.ThenTail's list; a card that matches but has
// since left the graveyard is not offered. Nothing to offer asks nothing.
func hobB9TakeFromMilled(ctx *Context, milled []uuid.UUID, match CardPredicate, question string, min, max int) error {
	player := ctx.Controller()
	var candidates []uuid.UUID
	for _, id := range milled {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !match(ctx.Game, player, c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: question,
		Cards:    candidates,
		Min:      min,
		Max:      max,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  hobB9MilledToHand,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// hobB9MilledToHand puts the picked cards from the graveyard into the hand,
// keyed (#4253): the asking item rides the prompt as data.
var hobB9MilledToHand = askedPick("choose-cards/hobbit-take-milled-into-hand", func(ctx *Context, r game.CardsChosen) error {
	return ReturnPickedToHandNow(ctx, r.Picked)
})
