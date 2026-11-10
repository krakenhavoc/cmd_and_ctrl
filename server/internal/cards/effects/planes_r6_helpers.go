package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_r6_helpers.go holds the shapes the redo-6 plane cards share
// (ADR 0216 slice 4, #4870). Append-only.

// planesR6APlayerCastsInstantOrSorcery is "Whenever a player casts an
// instant or sorcery spell": any player's cast, the plane's controller's
// included.
func planesR6APlayerCastsInstantOrSorcery(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (c.IsInstant() || c.IsSorcery())
}
