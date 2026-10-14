package effects

import (
	"testing"

	"github.com/google/uuid"
)

// A permanent the viewer does not know (a face-down one) carries neither
// list, cleared with the rest of the designations (ADR 0186 §6).
func TestCrewedByIsClearedForAViewerWhoDoesNotKnowTheCard(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	copter := pushVehicleForTest(g, owner.ID, "Smuggler's Copter", smugglersCopterOracle, 3, 3)
	crewer := pushCrewerForTest(g, owner.ID, "Crewer", 1)
	knownToTable(g, crewer)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == copter {
				g.Battlefield.Cards[i].KnownBy = map[uuid.UUID]bool{owner.ID: true}
			}
		}
	})
	crewBy(t, g, owner.ID, copter, crewer)

	if got := battlefieldViewCard(t, g, owner.ID.String(), copter).CrewedBy; len(got) != 1 {
		t.Fatalf("the knowing viewer's crewed_by = %v", got)
	}
	if got := battlefieldViewCard(t, g, opp.ID.String(), copter).CrewedBy; len(got) != 0 {
		t.Errorf("a viewer who does not know the card saw crewed_by %v", got)
	}
}
