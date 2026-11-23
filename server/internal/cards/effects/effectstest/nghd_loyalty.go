package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nghd_loyalty_test.go — a loyalty ability is an activated ability
// (CR 606.1), so naming a planeswalker with Phyrexian Revoker or Pithing
// Needle stops it, on the catalogued path and on the sandbox verb.

const NghdSorinLordOracle = "5c780dda-ee76-482a-a610-53c321c4fdec"

type NghdNamer func(*testing.T, *game.Game, uuid.UUID, string) uuid.UUID

func NghdNamedWalkerTable(t *testing.T, namer NghdNamer, named string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	namer(t, g, them.ID, named)
	sorin := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sorin, Lord of Innistrad", OracleID: NghdSorinLordOracle,
		TypeLine: "Legendary Planeswalker — Sorin", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	return g, me, sorin
}
