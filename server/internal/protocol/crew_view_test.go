package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// crew_view_test.go — the wire half of S27's Vehicle support. The
// client cannot build a crew prompt it can't see the number for, and
// it cannot offer the right creatures without the server saying which
// ones qualify. Both of those are one field each, and both are easy
// to get subtly wrong: the number because zero would erase it, the
// roster because the obvious filter (summoning sickness) is the one
// that must NOT be applied.

// seatCrewVehicle puts a Vehicle with a crew 3 ability on the
// battlefield under the first seat, with a roster of creatures
// around it.
func seatCrewVehicle(g *game.Game) (uuid.UUID, uuid.UUID) {
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Vehicle",
		TypeLine:   "Artifact — Vehicle",
		// Non-empty so stampActivatedAbilities doesn't skip it;
		// intrinsic abilities win over the catalog lookup.
		OracleID:   "00000000-0000-0000-0000-0000000000bb",
		Power:      4,
		Toughness:  4,
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "Crew 3", Cost: game.AbilityCost{Crew: 3}},
		},
	})
	return id, owner
}

func seatCreature(g *game.Game, owner uuid.UUID, name string, power int, tapped, sick bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:       id,
		Name:             name,
		TypeLine:         "Creature — Test",
		Power:            power,
		Toughness:        power,
		Owner:            owner,
		Controller:       owner,
		Tapped:           tapped,
		SummonedThisTurn: sick,
	})
	return id
}

func vehicleView(t *testing.T, g *game.Game, id uuid.UUID) CardView {
	t.Helper()
	v := ViewOfGame(g)
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("vehicle %s missing from the battlefield view", id)
	return CardView{}
}

func TestActivatedAbilityViewCarriesCrewCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	vehicle, owner := seatCrewVehicle(g)
	ready := seatCreature(g, owner, "Ready", 2, false, false)
	// Summoning-sick, and therefore STILL a legal crewer
	// (CR 702.122b) — tapping to crew is not paying a {T} cost.
	sick := seatCreature(g, owner, "Just cast", 2, false, true)
	// Tapped: not a legal crewer.
	seatCreature(g, owner, "Tapped", 5, true, false)
	// An opponent's creature: not a legal crewer.
	seatCreature(g, g.Seats[1].ID, "Theirs", 5, false, false)

	c := vehicleView(t, g, vehicle)
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(c.ActivatedAbilities))
	}
	ab := c.ActivatedAbilities[0]
	if ab.CrewCost != 3 {
		t.Errorf("crew_cost = %d, want 3", ab.CrewCost)
	}
	if ab.CrewOptions == nil {
		t.Fatal("crew_options is absent")
	}
	got := map[string]bool{}
	for _, id := range ab.CrewOptions.Cards {
		got[id] = true
	}
	if !got[ready.String()] {
		t.Error("an untapped creature is missing from crew_options")
	}
	if !got[sick.String()] {
		t.Error("a summoning-sick creature was filtered out of crew_options — CR 702.122b says it may crew")
	}
	if len(got) != 2 {
		t.Errorf("crew_options has %d cards, want 2 (a tapped creature and an opponent's are both out)", len(got))
	}
	if len(ab.CrewOptions.Players) != 0 {
		t.Error("crew_options offered players; crewing taps creatures")
	}
}

func TestNonCrewAbilityCarriesNoCrewFields(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Ordinary Artifact",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000cc",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "{T}: do a thing", Cost: game.AbilityCost{Tap: true}},
		},
	})
	c := vehicleView(t, g, id)
	ab := c.ActivatedAbilities[0]
	if ab.CrewCost != 0 || ab.CrewOptions != nil {
		t.Errorf("a non-crew ability carried crew fields: cost=%d options=%v", ab.CrewCost, ab.CrewOptions)
	}
}
