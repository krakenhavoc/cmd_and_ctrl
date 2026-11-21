package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2183 — the total life a player lost in ONE simultaneous event
// batch, read by Ob Nixilis, Captive Kingpin's "exactly 1 life".

const obNixilisKingpinOracle = "55b6434b-1542-40fd-b12a-697d40976582"

func ObKingpinBoard(t *testing.T) (g *game.Game, me *game.Player, ob uuid.UUID, top uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me = g.Seats[0]
	AdvanceToMain(t, g)
	ob = PushCatalogPermanent(g, me.ID, "Ob Nixilis, Captive Kingpin", "Legendary Creature — Demon", obNixilisKingpinOracle, false)
	top = SeedLibrary(me, "Top Card")[0]
	return g, me, ob, top
}

func SettleLifeBoundary(g *game.Game) { g.RunStateChecksForTest() }
