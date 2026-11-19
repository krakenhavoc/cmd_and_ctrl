package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cardsToYourGraveyardNotFromBattlefield counts the cards `owner` had put
// into their graveyard this turn from anywhere other than the battlefield:
// discards, mills, surveils, spells that resolved or were countered, cards
// pulled out of exile. It reads the turn's events (graveyardArrivals, one
// arrival per card), so a card that has since left the graveyard still
// counts and a token never does (CR 108.2b).
func cardsToYourGraveyardNotFromBattlefield(g *game.Game, owner uuid.UUID) int {
	arrivals := graveyardArrivals{Except: []game.ZoneKind{game.ZoneBattlefield}}
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if arrivals.arrival(ev, g, owner) {
			n++
		}
	}
	return n
}

// CondThreeCardsToYourGraveyardNotFromBattlefield is "three or more cards
// were put into your graveyard from anywhere other than the battlefield
// this turn" (Dimir Strandcatcher), an intervening if (CR 603.4).
type CondThreeCardsToYourGraveyardNotFromBattlefield struct{}

func (CondThreeCardsToYourGraveyardNotFromBattlefield) Holds(ctx *Context) bool {
	return cardsToYourGraveyardNotFromBattlefield(ctx.Game, ctx.Controller()) >= 3
}

func (CondThreeCardsToYourGraveyardNotFromBattlefield) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondGraveyard, 0), true
}
