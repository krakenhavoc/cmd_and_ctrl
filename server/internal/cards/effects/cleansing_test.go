package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cleansingOracle = "46b83657-1699-4162-a4bb-b792da396fe6"

func cleansingLand(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return seedPermanentWithOracle(g, owner, name, "Basic Land — "+name, "")
}

// cleansingCast casts Cleansing and returns once the first leg is open.
func cleansingCast(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Cleansing", "Sorcery", cleansingOracle, nil)
	passPriorityAroundTable(t, g)
}

// cleansingAnswer answers every open leg: picks[chooser] is that player's
// choice (nil = pay for nothing).
func cleansingAnswer(t *testing.T, g *game.Game, picks map[uuid.UUID][]uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		leg := eachOpenLeg(g)
		if leg == nil {
			return
		}
		if err := g.ResolveChooseCards(leg.ID, leg.Chooser, picks[leg.Chooser]); err != nil {
			t.Fatalf("answer for %v: %v", leg.Chooser, err)
		}
	}
	t.Fatal("legs never closed")
}

func TestCleansingDestroysEveryLandWhenNobodyPays(t *testing.T) {
	g := newCatalogGame(t)
	var lands []uuid.UUID
	for i, name := range []string{"Forest", "Island", "Swamp", "Plains"} {
		lands = append(lands, cleansingLand(g, g.Seats[i].ID, name))
	}
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	cleansingCast(t, g)
	if eachOpenLeg(g) == nil {
		t.Fatal("no per-land offer is open")
	}
	cleansingAnswer(t, g, nil)
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
	g := newCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	a1 := cleansingLand(g, a.ID, "Forest")
	a2 := cleansingLand(g, a.ID, "Island")
	b1 := cleansingLand(g, b.ID, "Swamp")
	cleansingCast(t, g)
	cleansingAnswer(t, g, map[uuid.UUID][]uuid.UUID{a.ID: {a1, a2}})
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
	g := newCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	shared := cleansingLand(g, a.ID, "Forest")
	other := cleansingLand(g, b.ID, "Island")
	cleansingCast(t, g)
	cleansingAnswer(t, g, map[uuid.UUID][]uuid.UUID{a.ID: {shared}, b.ID: {shared}})
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
	g := newCatalogGame(t)
	for i, name := range []string{"Forest", "Island", "Swamp"} {
		cleansingLand(g, g.Seats[i].ID, name)
	}
	cleansingCast(t, g)
	cleansingAnswer(t, g, nil)
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
	g := newCatalogGame(t)
	cit := seedPermanentWithOracle(g, g.Seats[0].ID, "Darksteel Citadel", "Artifact Land", b02bDarksteelCitadelOracle)
	plain := cleansingLand(g, g.Seats[1].ID, "Island")
	cleansingCast(t, g)
	cleansingAnswer(t, g, nil)
	if !g.Battlefield.Contains(cit) {
		t.Error("an indestructible land was destroyed")
	}
	if g.Battlefield.Contains(plain) {
		t.Error("an ordinary unpaid land survived")
	}
}
