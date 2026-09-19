package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_then_take.go — "mill N cards. You may put a <kind> card from
// among them into your hand", and its wider sibling that ranges over
// the whole graveyard.
//
// Two shapes, and the printed wording is the difference:
//
//   - "from among THEM" (Barrowgoyf, Six) offers only the cards this
//     mill put there. A card that was already in the graveyard is not
//     among them, and neither is one the mill sent somewhere else —
//     a commander that answered CR 903.9, or a card an "if it would
//     be put into a graveyard, exile it instead" replacement caught.
//     That is why the candidate list is built from MillToZone.Then's
//     `milled` and re-checked against the graveyard rather than
//     re-derived from the pile.
//   - "from YOUR GRAVEYARD" (Overlord of the Balemurk) offers the
//     whole pile, the four cards just milled included. The mill still
//     has to finish first, which is why it is written inside the
//     mill's continuation and not on the line after it.
//
// Both are "you MAY", so both floor the pick at zero: declining is
// always an answer, and a matching card that is not there is not an
// error (CR 608.2c).

// mayTakeOneFromAmongThem is "you may put a <match> card from among
// them into your hand", asked of the resolving effect's controller.
//
// `milled` is MillToZone.Then's list, in library order. Candidates
// are filtered to cards that both match and are still in the
// controller's graveyard — the second half matters, because anything
// could have moved a card out of the pile between the mill and this
// prompt, and offering it would queue a pick the submit-side zone
// check then refuses.
func mayTakeOneFromAmongThem(ctx *Context, milled []uuid.UUID, match CardPredicate, question string) error {
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
	return mayTakeOneIntoYourHand(ctx, candidates, question)
}

// mayTakeOneFromYourGraveyard is "you may return a <match> card from
// your graveyard to your hand" — the same prompt over the whole pile
// rather than over one mill's output.
func mayTakeOneFromYourGraveyard(ctx *Context, match CardPredicate, question string) error {
	player := ctx.Controller()
	p := ctx.PlayerByID(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match(ctx.Game, player, c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	return mayTakeOneIntoYourHand(ctx, candidates, question)
}

// mayTakeOneIntoYourHand queues the shared prompt: pick at most one of
// `candidates` and move it from the graveyard to its owner's hand.
//
// Nothing to offer queues nothing. A floor of zero with an empty
// candidate list would be a prompt whose only answer is "none", which
// is a click the player had no choice about and, on a card that mills
// every combat, one per attack.
//
// The Then closure captures the stack item and rebuilds its Context
// from the game the resolver hands back, the contract every queued
// continuation follows: an undo restores a different *Game, and a
// closure holding the old one would move a card in a game nobody is
// looking at.
func mayTakeOneIntoYourHand(ctx *Context, candidates []uuid.UUID, question string) error {
	if len(candidates) == 0 {
		return nil
	}
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: question,
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneGraveyard,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ReturnFromGraveyard{Target: picked[0], Dest: game.ZoneHand}.Apply(NewContext(g, item))
		},
	})
	return nil
}
