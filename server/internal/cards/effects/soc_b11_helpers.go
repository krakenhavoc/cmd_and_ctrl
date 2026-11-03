package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b11_helpers.go — shared pieces of Secrets of Strixhaven Commander
// batch 11 (#4834): the two Quintorius cards that make a Spirit when cards
// leave a graveyard, and the Spirit-lord half they share.

// socB11CardLeftYourGraveyard is "one or more cards leave your graveyard"
// for one zone-move event: any card its controller owns, out of a
// graveyard and into any other zone. Desecrated Tomb's creature-only
// check (b02CreatureCardLeftYourGraveyard) without the type.
func socB11CardLeftYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventZoneMove || ev.OldZone != game.ZoneGraveyard || ev.NewZone == game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// socB11SpiritWhenCardsLeaveYourGraveyard is the ability both
// Quintorius, Field Historian and Quintorius, History Chaser print:
// "Whenever one or more cards leave your graveyard, create a 3/2 red and
// white Spirit creature token." A batch trigger (OncePerBatch): cards that
// leave as one event make one Spirit, cards that leave as separate events
// make one each (the 2021-04-16 ruling).
func socB11SpiritWhenCardsLeaveYourGraveyard(name string) game.TriggeredAbility {
	return OncePerBatch(On(game.EventZoneMove, socB11CardLeftYourGraveyard,
		name+" — create a 3/2 red and white Spirit creature token",
		Do(CreateToken{Template: TokenCard("3/2 red and white Spirit"), N: 1})))
}

// socB11QuintoriusMillDraw is "draw two cards, then mill a card" after the
// discard of Quintorius, History Chaser's +1, as registered continuations
// (a table waiting on the discard or on a draw replacement is a restore
// point). The mill is the draw's own "then", so it waits for any prompt the
// draws pause on.
var (
	socB11QuintoriusMill = game.RegisterDrawThen("draw-then/quintorius-history-chaser-mill", func(g *game.Game, d game.DrawThen) error {
		return g.MillNForEffect(d.Player, 1)
	})
	socB11QuintoriusDiscardDraw = discardDraw{then: game.RegisterDiscardThen("choose-cards/quintorius-history-chaser-draw-two-mill-one",
		func(g *game.Game, r game.DiscardsDone) error {
			if r.Discarded.Count() == 0 {
				return nil
			}
			return g.DrawNThenForEffect(r.Player, 2, game.DrawThen{Ref: socB11QuintoriusMill, Player: r.Player})
		})}
)
