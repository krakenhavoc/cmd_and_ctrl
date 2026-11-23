package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exileThisThenScheduleReturn is "Exile this <permanent>. Return it to the
// battlefield under its owner's control at the beginning of the next end
// step" (Djinn of the Fountain, Stenn, Paranoid Partisan). Only the
// permanent the ability came from is exiled: one that left, or came back as
// a new object, in response has nothing to blink (CR 400.7, CR 608.2b).
func ExileThisThenScheduleReturn(ctx *Context, item *game.StackItem, label string) error {
	if !SourceIsStillThisPermanent(ctx.Game, item) {
		return nil
	}
	return ExileCardsThenScheduleReturn(ctx, []uuid.UUID{item.SourceCardID}, label)
}

// CastFromGraveyardByDiscarding is "You may cast this card from your
// graveyard by discarding <n> cards in addition to paying its other costs"
// (Rona, Sheoldred's Faithful): Retrace's model with the land card widened
// to n cards of any kind, and Quilled Greatwurm's claim-opens-the-zone
// shape. The mana is the printed cost, spelled as the card prints it; the
// discard is a cost (CR 601.2h) paid with the spell on the stack, so a
// countered Rona does not give the cards back. There is no exile clause: a
// Rona that dies goes back to the graveyard, ready to be cast again.
//
// The card file still opens the zone:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{CastFromGraveyardByDiscarding("{1}{U}{B}{B}", 2)},
func CastFromGraveyardByDiscarding(printedCost string, n int) game.AlternativeCost {
	what := NumberWord(n) + " cards"
	return game.AlternativeCost{
		Key:             "discard_from_graveyard",
		Label:           "Discard " + what + " to cast this card from your graveyard",
		ManaCost:        printedCost,
		FromZone:        game.ZoneGraveyard,
		DiscardFromHand: CardInYourHand(what).WithCount(n, n),
		PayLabel:        what,
	}
}
