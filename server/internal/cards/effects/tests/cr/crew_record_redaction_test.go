package testscr

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// A permanent the viewer does not know (a face-down one) carries neither
// list, cleared with the rest of the designations (ADR 0186 §6).
func TestCrewedByIsClearedForAViewerWhoDoesNotKnowTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	copter := effectstest.PushVehicleForTest(g, owner.ID, "Smuggler's Copter", effectstest.SmugglersCopterOracle, 3, 3)
	crewer := effectstest.PushCrewerForTest(g, owner.ID, "Crewer", 1)
	effectstest.KnownToTable(g, crewer)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == copter {
				g.Battlefield.Cards[i].KnownBy = map[uuid.UUID]bool{owner.ID: true}
			}
		}
	})
	effectstest.CrewBy(t, g, owner.ID, copter, crewer)

	if got := battlefieldViewCard(t, g, owner.ID.String(), copter).CrewedBy; len(got) != 1 {
		t.Fatalf("the knowing viewer's crewed_by = %v", got)
	}
	if got := battlefieldViewCard(t, g, opp.ID.String(), copter).CrewedBy; len(got) != 0 {
		t.Errorf("a viewer who does not know the card saw crewed_by %v", got)
	}
}
