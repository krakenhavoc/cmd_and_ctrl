package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_play_ordinal.go — Fastbond's "Whenever you play a land, if it
// wasn't the first land you played this turn" (#6213; CR 305.2).

// LandPlayedAfterTheFirst reports whether ev is `player`'s land PLAY
// (CR 305.1, Event.Played) that wasn't the first land they played this
// turn. It reads the play's own ordinal (Event.LandOrdinal), so the
// answer does not move with the turn's tally after the land landed.
func LandPlayedAfterTheFirst(ev game.Event, player uuid.UUID) bool {
	return ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield && ev.Played &&
		ev.Actor == player && ev.LandOrdinal > 1
}
