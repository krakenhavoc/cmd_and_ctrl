package game

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

func spawnTestGame(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()
	g := NewGame()
	alice, err := g.AddPlayer("Alice", buildTestDeck("Atraxa, Praetors' Voice"))
	if err != nil {
		t.Fatalf("AddPlayer alice: %v", err)
	}
	bob, err := g.AddPlayer("Bob", buildTestDeck("Atraxa, Praetors' Voice"))
	if err != nil {
		t.Fatalf("AddPlayer bob: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, alice, bob
}

func devTemplate() Card {
	return Card{
		Name:       "Lightning Bolt",
		ScryfallID: "77c6fa74-5543-42ac-9ead-0e890b188e99",
		OracleID:   "4457ed35-7c10-48c8-9776-456485fdf070",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
	}
}

func TestSpawnCardsForDevZoneRouting(t *testing.T) {
	for _, tc := range []struct {
		zone ZoneKind
		size func(g *Game, p *Player) int
	}{
		{ZoneBattlefield, func(g *Game, _ *Player) int { return g.Battlefield.Size() }},
		{ZoneExile, func(g *Game, _ *Player) int { return g.Exile.Size() }},
		{ZoneHand, func(_ *Game, p *Player) int { return p.Hand.Size() }},
		{ZoneGraveyard, func(_ *Game, p *Player) int { return p.Graveyard.Size() }},
		{ZoneLibrary, func(_ *Game, p *Player) int { return p.Library.Size() }},
		{ZoneCommand, func(_ *Game, p *Player) int { return p.Command.Size() }},
	} {
		t.Run(string(tc.zone), func(t *testing.T) {
			g, alice, _ := spawnTestGame(t)
			before := tc.size(g, alice)
			ids, err := g.SpawnCardsForDev(alice.ID, tc.zone, devTemplate(), 2)
			if err != nil {
				t.Fatalf("SpawnCardsForDev: %v", err)
			}
			if len(ids) != 2 {
				t.Fatalf("got %d ids, want 2", len(ids))
			}
			if got := tc.size(g, alice); got != before+2 {
				t.Errorf("zone size = %d, want %d", got, before+2)
			}
			if ids[0] == ids[1] {
				t.Error("spawned cards share an InstanceID")
			}
		})
	}
}

// A spawned card must be indistinguishable from an imported one, or
// the effect catalog (which keys on OracleID) won't match it and the
// tool silently tests the wrong thing.
func TestSpawnCardsForDevPreservesIdentity(t *testing.T) {
	g, alice, _ := spawnTestGame(t)
	tmpl := devTemplate()
	if _, err := g.SpawnCardsForDev(alice.ID, ZoneBattlefield, tmpl, 1); err != nil {
		t.Fatalf("SpawnCardsForDev: %v", err)
	}
	got, err := g.Battlefield.Top()
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if got.OracleID != tmpl.OracleID || got.ScryfallID != tmpl.ScryfallID {
		t.Errorf("identity not preserved: oracle=%q scryfall=%q", got.OracleID, got.ScryfallID)
	}
	if got.Controller != alice.ID || got.Owner != alice.ID {
		t.Errorf("controller/owner = %v/%v, want %v", got.Controller, got.Owner, alice.ID)
	}
	if got.InstanceID == uuid.Nil || got.InstanceID == tmpl.InstanceID {
		t.Error("spawned card did not get a fresh InstanceID")
	}
}

func TestSpawnCardsForDevVisibility(t *testing.T) {
	g, alice, bob := spawnTestGame(t)

	if _, err := g.SpawnCardsForDev(alice.ID, ZoneHand, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to hand: %v", err)
	}
	inHand, _ := alice.Hand.Top()
	if !inHand.IsKnownTo(alice.ID) {
		t.Error("owner should know a card spawned into their own hand")
	}
	if inHand.IsKnownTo(bob.ID) {
		t.Error("opponent must NOT know a card spawned into another player's hand")
	}

	if _, err := g.SpawnCardsForDev(alice.ID, ZoneBattlefield, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to battlefield: %v", err)
	}
	onBF, _ := g.Battlefield.Top()
	if !onBF.IsKnownTo(alice.ID) || !onBF.IsKnownTo(bob.ID) {
		t.Error("every seat should know a card spawned onto the battlefield")
	}

	if _, err := g.SpawnCardsForDev(alice.ID, ZoneLibrary, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to library: %v", err)
	}
	inLib, _ := alice.Library.Top()
	if inLib.IsKnownTo(alice.ID) || inLib.IsKnownTo(bob.ID) {
		t.Error("a card spawned into a library should be known to nobody")
	}
}

// Battlefield spawns must emit ETB or the tool can't be used to test
// the interactions it exists for.
func TestSpawnCardsForDevEmitsETB(t *testing.T) {
	g, alice, _ := spawnTestGame(t)
	before := len(g.Events)
	ids, err := g.SpawnCardsForDev(alice.ID, ZoneBattlefield, devTemplate(), 1)
	if err != nil {
		t.Fatalf("SpawnCardsForDev: %v", err)
	}
	var sawETB bool
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventETB && ev.CardID == ids[0] {
			sawETB = true
		}
	}
	if !sawETB {
		t.Error("battlefield spawn did not emit EventETB")
	}

	// A spawn into hand must NOT emit ETB.
	before = len(g.Events)
	if _, err := g.SpawnCardsForDev(alice.ID, ZoneHand, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to hand: %v", err)
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventETB {
			t.Error("hand spawn emitted EventETB")
		}
	}
}

func TestSpawnCardsForDevRejects(t *testing.T) {
	g, alice, _ := spawnTestGame(t)

	for _, n := range []int{0, -1, MaxDevSpawnCount + 1} {
		if _, err := g.SpawnCardsForDev(alice.ID, ZoneHand, devTemplate(), n); err != ErrDevSpawnCount {
			t.Errorf("count %d: err = %v, want ErrDevSpawnCount", n, err)
		}
	}
	if _, err := g.SpawnCardsForDev(alice.ID, ZoneStack, devTemplate(), 1); err != ErrDevZoneUnsupported {
		t.Errorf("stack: err = %v, want ErrDevZoneUnsupported", err)
	}
	if _, err := g.SpawnCardsForDev(uuid.New(), ZoneHand, devTemplate(), 1); err != ErrPlayerNotFound {
		t.Errorf("unseated player: err = %v, want ErrPlayerNotFound", err)
	}
}
