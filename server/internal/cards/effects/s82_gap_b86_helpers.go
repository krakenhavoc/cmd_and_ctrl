package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapB86HistoricPlayedFromYourLibrary is The Fourth Doctor's "when you do": its
// controller cast a historic spell from their library, or played (CR 305.4) a
// historic land from it.
func gapB86HistoricPlayedFromYourLibrary(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventCast:
		if ev.Actor != source.Controller || ev.OldZone != game.ZoneLibrary {
			return false
		}
	case game.EventZoneMove:
		if ev.Actor != source.Controller || !ev.Played || ev.OldZone != game.ZoneLibrary {
			return false
		}
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && b09IsHistoric(c)
}
