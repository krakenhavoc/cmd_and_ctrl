package protocol

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saddle_view_test.go — #2695, CR 702.171: the wire half of saddle.
// The saddle ability rides crew's two fields (crew_cost, crew_options)
// plus a `saddle` flag, and the Mount's designation is `saddled`.

// A Mount's saddle ability carries the number, the flag, and a roster
// that leaves the Mount itself out — "other" creatures.
func TestSaddleAbilityViewCarriesNumberFlagAndOtherCreaturesOnly(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	mount := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: mount, Name: "Test Mount", TypeLine: "Creature — Horse Mount",
		OracleID: "00000000-0000-0000-0000-0000000000dd", Power: 5, Toughness: 5,
		Owner: owner, Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "Saddle 2", Cost: game.AbilityCost{Saddle: 2}, SorcerySpeed: true},
		},
	})
	other := seatCreature(g, owner, "Saddler", 2, false, true) // summoning-sick may still saddle
	seatCreature(g, owner, "Tapped", 5, true, false)
	seatCreature(g, g.Seats[1].ID, "Theirs", 5, false, false)

	c := vehicleView(t, g, mount)
	if len(c.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(c.ActivatedAbilities))
	}
	ab := c.ActivatedAbilities[0]
	if ab.CrewCost != 2 || !ab.Saddle {
		t.Errorf("crew_cost = %d, saddle = %v; want 2 and true", ab.CrewCost, ab.Saddle)
	}
	if ab.CrewOptions == nil {
		t.Fatal("crew_options is absent")
	}
	if len(ab.CrewOptions.Cards) != 1 || ab.CrewOptions.Cards[0] != other.String() {
		t.Errorf("crew_options = %v, want only the one other untapped creature (not the Mount)", ab.CrewOptions.Cards)
	}
}

// A crew ability is not a saddle ability: the flag is absent.
func TestCrewAbilityViewDoesNotCarryTheSaddleFlag(t *testing.T) {
	g := buildActiveGame(t)
	vehicle, _ := seatCrewVehicle(g)
	if c := vehicleView(t, g, vehicle); c.ActivatedAbilities[0].Saddle {
		t.Error("a crew ability projects saddle")
	}
}

// CardView.saddled is public on the battlefield, like monstrous, and
// absent for a Mount that is not saddled.
func TestSaddledViewIsPublicOnTheBattlefield(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	saddledID, plainID := uuid.New(), uuid.New()
	now := time.Now().UnixNano()
	seen := map[uuid.UUID]bool{owner.ID: true, g.Seats[1].ID: true}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: saddledID, Name: "Saddled Mount", TypeLine: "Creature — Horse Mount",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, Saddled: true, KnownBy: seen,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: plainID, Name: "Plain Mount", TypeLine: "Creature — Horse Mount",
			Owner: owner.ID, Controller: owner.ID, EnteredBattlefieldAt: now, KnownBy: seen,
		})
	})
	v := ViewOfGameFor(g, g.Seats[1].ID.String())
	got := map[string]bool{}
	for _, c := range v.Battlefield.Cards {
		got[c.InstanceID] = c.Saddled
	}
	if !got[saddledID.String()] {
		t.Error("a saddled Mount does not project saddled to an opponent")
	}
	if got[plainID.String()] {
		t.Error("a Mount that is not saddled projects saddled")
	}
}
