package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guffc_test.go — the Guffster's Party deck request (#4184), slice c:
// Onakke Oathkeeper, Teyo, Sarkhan the Masterless, Narset, Enlightened
// Master, Vronos, Final Showdown and Guff Rewrites History.

const (
	GuffcOnakkeOracle = "080df9eb-c26b-494a-96b4-d7e47e501866"
	GuffcTeyoOracle   = "0dfdf1a1-a3a9-490f-bd9d-dcba088c6785"
	GuffcSarkhan      = "67dbc5dd-6d05-43e7-83f7-70e10eb54807"
	GuffcNarsetOracle = "aeaf7768-ed67-487d-8506-e667babb88ba"
	GuffcVronosOracle = "86851bdb-9b2f-4407-97a1-a6a9f2e0ef66"
	GuffcShowdown     = "7e7ec3d6-a84f-4cc3-93f4-4d181d41e126"
	GuffcRewrites     = "7ff7af5f-e4e3-40af-8882-756564bfa479"
)

func GuffcWalker(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	t.Helper()
	id := PushCatalogPermanent(g, owner, name, typeLine, oracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	return id
}
