package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// station_ability_test.go — #759: the station ABILITY (CR 702.184a),
// Station(), on #758's tap-another cost, and the four proof cards that
// carry it. The threshold half is station_test.go.

const (
	GalvanizingSawshipOracle  = "dfe8f77a-cc26-438b-92ae-2ca7a91f813b"
	UthrosResearchCraftOracle = "e1c9783a-1d1b-40d7-872e-0ca11b229ce6"
	AdagiaOracle              = "70d35dbd-1d91-4a2a-a643-6870d168f4f5"
)

// stationBoard seats a station permanent and one creature of the given
// power under the active seat, at its precombat main phase, and
// returns the seat and both IDs.
func StationBoard(t *testing.T, name, typeLine, oracle string, power int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ship := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	crew := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Station Hand",
		TypeLine:   "Creature — Human Pilot",
		Power:      power,
		Toughness:  3,
		Owner:      me.ID,
		Controller: me.ID,
	})
	return g, me, ship, crew
}

// stationIndex is the position of the station ability on a card, found
// by its label so a card that declares others first still works.
func StationIndex(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c := LayeredCard(t, g, id)
	for i, a := range game.ActivatedAbilitiesForCard(c) {
		if a.Label == effects.StationLabel {
			return i
		}
	}
	t.Fatalf("%s has no station ability", c.Name)
	return -1
}

func Station(t *testing.T, g *game.Game, me *game.Player, ship, crew uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, ship, StationIndex(t, g, ship), game.ActivateAbilityParams{TapIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("station: %v", err)
	}
}
