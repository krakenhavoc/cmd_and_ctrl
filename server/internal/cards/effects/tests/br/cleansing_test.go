package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCleansingDestroysEveryLandWhenNobodyPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	var lands []uuid.UUID
	for i, name := range []string{"Forest", "Island", "Swamp", "Plains"} {
		lands = append(lands, effectstest.CleansingLand(g, g.Seats[i].ID, name))
	}
	bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CleansingCast(t, g)
	if effectstest.EachOpenLeg(g) == nil {
		t.Fatal("no per-land offer is open")
	}
	effectstest.CleansingAnswer(t, g, nil)
	for _, id := range lands {
		if g.Battlefield.Contains(id) {
			t.Errorf("land %v survived with nobody paying", id)
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature is not a land")
	}
	for _, s := range g.Seats {
		if s.Life != game.StartingLife {
			t.Errorf("%s lost life without paying", s.Name)
		}
	}
}

func TestCleansingPaidLandSurvivesAndEachPayerLosesPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	a1 := effectstest.CleansingLand(g, a.ID, "Forest")
	a2 := effectstest.CleansingLand(g, a.ID, "Island")
	b1 := effectstest.CleansingLand(g, b.ID, "Swamp")
	effectstest.CleansingCast(t, g)
	effectstest.CleansingAnswer(t, g, map[uuid.UUID][]uuid.UUID{a.ID: {a1, a2}})
	if !g.Battlefield.Contains(a1) || !g.Battlefield.Contains(a2) {
		t.Error("a paid-for land was destroyed")
	}
	if g.Battlefield.Contains(b1) {
		t.Error("an unpaid land survived")
	}
	if a.Life != game.StartingLife-2 {
		t.Errorf("seat 0 life = %d, want %d (2 lands)", a.Life, game.StartingLife-2)
	}
	if b.Life != game.StartingLife {
		t.Errorf("seat 1 life = %d, paid nothing", b.Life)
	}
}

func TestCleansingTwoPlayersOnOneLandBothPayAndItSurvives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	shared := effectstest.CleansingLand(g, a.ID, "Forest")
	other := effectstest.CleansingLand(g, b.ID, "Island")
	effectstest.CleansingCast(t, g)
	effectstest.CleansingAnswer(t, g, map[uuid.UUID][]uuid.UUID{a.ID: {shared}, b.ID: {shared}})
	if !g.Battlefield.Contains(shared) {
		t.Error("the doubly paid land was destroyed")
	}
	if g.Battlefield.Contains(other) {
		t.Error("an unpaid land survived")
	}
	if a.Life != game.StartingLife-1 || b.Life != game.StartingLife-1 {
		t.Errorf("lives = %d, %d; both payers owe 1", a.Life, b.Life)
	}
}

func TestCleansingUnpaidLandsDieInOneBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for i, name := range []string{"Forest", "Island", "Swamp"} {
		effectstest.CleansingLand(g, g.Seats[i].ID, name)
	}
	effectstest.CleansingCast(t, g)
	effectstest.CleansingAnswer(t, g, nil)
	batches := map[uint64]int{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventLTB {
			batches[ev.Batch]++
		}
	}
	if len(batches) != 1 {
		t.Fatalf("leaves-the-battlefield events fell in %d batches %v, want one", len(batches), batches)
	}
	for _, n := range batches {
		if n != 3 {
			t.Errorf("batch holds %d events, want 3", n)
		}
	}
}

func TestCleansingIndestructibleLandSurvivesAnyway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	cit := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, "Darksteel Citadel", "Artifact Land", effectstest.B02bDarksteelCitadelOracle)
	plain := effectstest.CleansingLand(g, g.Seats[1].ID, "Island")
	effectstest.CleansingCast(t, g)
	effectstest.CleansingAnswer(t, g, nil)
	if !g.Battlefield.Contains(cit) {
		t.Error("an indestructible land was destroyed")
	}
	if g.Battlefield.Contains(plain) {
		t.Error("an ordinary unpaid land survived")
	}
}
