package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vehicles_test.go — S27, CR 301.7 and CR 702.122.
//
// Crew is the first cost in the engine paid by tapping permanents
// OTHER than the source, and the first counted against a total rather
// than an item count. The tests below are about exactly those two
// differences, plus the one rule that is easiest to get wrong in the
// strict direction: a creature that arrived this turn may crew.

const (
	SmugglersCopterOracle  = "49136bdc-bc50-49a2-999a-1ef9c16ea130"
	EsikasChariotOracle    = "8e7b079d-9ede-421c-bd2b-f9a5126a8e6f"
	CultivatorsCaravanOrcl = "c1eb530c-dd36-40ae-8617-6bb6969565e1"
)

// pushVehicleForTest puts a Vehicle on the battlefield with its
// printed P/T, the way the deck importer stamps one. A Vehicle's
// power and toughness are printed card data like any creature's —
// what it lacks until it is crewed is the CREATURE type.
func PushVehicleForTest(g *game.Game, owner uuid.UUID, name, oracleID string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   "Artifact — Vehicle",
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// pushCrewerForTest puts an untapped creature with the given power on
// the battlefield, already past its summoning sickness so tests that
// are not about sickness don't have to think about it.
func PushCrewerForTest(g *game.Game, owner uuid.UUID, name string, power int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      power,
		Toughness:  power,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// advancePastCleanupOf walks the turn engine forward until `seat` is
// no longer the active player — which necessarily crosses that
// seat's cleanup step, where CR 514.2 durations expire.
func AdvancePastCleanupOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep past cleanup of seat %d: %v", seat, err)
		}
		if g.Turn.ActiveSeat != seat {
			return
		}
	}
	t.Fatalf("seat %d never gave up the turn", seat)
}

func BattlefieldCardByID(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// crewForTest pays "Crew 1" with one creature and settles the
// ability, so a test that is about what the crewed Vehicle then does
// starts from a Vehicle that is already a creature.
func CrewForTest(t *testing.T, g *game.Game, controller, vehicle, crewer uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, vehicle, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{crewer},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	PassPriorityAroundTable(t, g)
	v, ok := BattlefieldCardByID(g, vehicle)
	if !ok || !v.IsCreature() {
		t.Fatalf("the crewed Vehicle is not a creature")
	}
}
