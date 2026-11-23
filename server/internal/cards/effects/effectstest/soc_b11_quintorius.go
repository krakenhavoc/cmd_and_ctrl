package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Quintorius, History Chaser ---

func SocB11Chaser(g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	q := PushCatalogPermanent(g, me.ID, "Quintorius, History Chaser", "Legendary Planeswalker — Quintorius", SocB11HistoryChaserOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, loyalty) })
	return q
}
