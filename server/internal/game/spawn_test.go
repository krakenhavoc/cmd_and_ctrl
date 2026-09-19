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

func TestSpawnCardsZoneRouting(t *testing.T) {
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
			ids, err := g.SpawnCards(uuid.Nil, alice.ID, tc.zone, devTemplate(), 2)
			if err != nil {
				t.Fatalf("SpawnCards: %v", err)
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
func TestSpawnCardsPreservesIdentity(t *testing.T) {
	g, alice, _ := spawnTestGame(t)
	tmpl := devTemplate()
	if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneBattlefield, tmpl, 1); err != nil {
		t.Fatalf("SpawnCards: %v", err)
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

func TestSpawnCardsVisibility(t *testing.T) {
	g, alice, bob := spawnTestGame(t)

	if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneHand, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to hand: %v", err)
	}
	inHand, _ := alice.Hand.Top()
	if !inHand.IsKnownTo(alice.ID) {
		t.Error("owner should know a card spawned into their own hand")
	}
	if inHand.IsKnownTo(bob.ID) {
		t.Error("opponent must NOT know a card spawned into another player's hand")
	}

	if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneBattlefield, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to battlefield: %v", err)
	}
	onBF, _ := g.Battlefield.Top()
	if !onBF.IsKnownTo(alice.ID) || !onBF.IsKnownTo(bob.ID) {
		t.Error("every seat should know a card spawned onto the battlefield")
	}

	if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneLibrary, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to library: %v", err)
	}
	inLib, _ := alice.Library.Top()
	if inLib.IsKnownTo(alice.ID) || inLib.IsKnownTo(bob.ID) {
		t.Error("a card spawned into a library should be known to nobody")
	}
}

// Battlefield spawns must emit ETB or the tool can't be used to test
// the interactions it exists for.
func TestSpawnCardsEmitsETB(t *testing.T) {
	g, alice, _ := spawnTestGame(t)
	before := len(g.Events)
	ids, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneBattlefield, devTemplate(), 1)
	if err != nil {
		t.Fatalf("SpawnCards: %v", err)
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
	if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneHand, devTemplate(), 1); err != nil {
		t.Fatalf("spawn to hand: %v", err)
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventETB {
			t.Error("hand spawn emitted EventETB")
		}
	}
}

func TestSpawnCardsRejects(t *testing.T) {
	g, alice, _ := spawnTestGame(t)

	for _, n := range []int{0, -1, MaxSpawnCount + 1} {
		if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneHand, devTemplate(), n); err != ErrSpawnCount {
			t.Errorf("count %d: err = %v, want ErrSpawnCount", n, err)
		}
	}
	if _, err := g.SpawnCards(uuid.Nil, alice.ID, ZoneStack, devTemplate(), 1); err != ErrSpawnZoneUnsupported {
		t.Errorf("stack: err = %v, want ErrSpawnZoneUnsupported", err)
	}
	if _, err := g.SpawnCards(uuid.Nil, uuid.New(), ZoneHand, devTemplate(), 1); err != ErrPlayerNotFound {
		t.Errorf("unseated player: err = %v, want ErrPlayerNotFound", err)
	}
}

// ADR 0075 §2.4: every spawn is announced. ONE event per request,
// carrying who asked, whose zone it landed in, what arrived and how
// many — the projection needs all four to write the line, and a
// per-card event would write it n times.
func TestSpawnCardsAnnouncesOnce(t *testing.T) {
	g, alice, bob := spawnTestGame(t)

	before := len(g.Events)
	if _, err := g.SpawnCards(bob.ID, alice.ID, ZoneGraveyard, devTemplate(), 3); err != nil {
		t.Fatalf("SpawnCards: %v", err)
	}

	var spawned []Event
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventSpawned {
			spawned = append(spawned, ev)
		}
	}
	if len(spawned) != 1 {
		t.Fatalf("emitted %d EventSpawned, want exactly 1", len(spawned))
	}
	ev := spawned[0]
	if ev.Actor != bob.ID {
		t.Errorf("Actor = %v, want the spawner %v", ev.Actor, bob.ID)
	}
	if ev.Target != alice.ID {
		t.Errorf("Target = %v, want the controller %v", ev.Target, alice.ID)
	}
	if ev.NewZone != ZoneGraveyard || ev.Label != "Lightning Bolt" || ev.Amount != 3 {
		t.Errorf("event = %+v, want 3 Lightning Bolt into the graveyard", ev)
	}
}

// The announcement precedes the cards, so the log reads in the order
// the table watched it: "Bob spawned Lightning Bolt", then whatever
// the card entering set off.
func TestSpawnAnnouncementPrecedesTheEntry(t *testing.T) {
	g, alice, _ := spawnTestGame(t)
	before := len(g.Events)
	if _, err := g.SpawnCards(alice.ID, alice.ID, ZoneBattlefield, devTemplate(), 1); err != nil {
		t.Fatalf("SpawnCards: %v", err)
	}
	announced, entered := -1, -1
	for i, ev := range g.Events[before:] {
		switch {
		case ev.Kind == EventSpawned && announced < 0:
			announced = i
		case ev.Kind == EventETB && entered < 0:
			entered = i
		}
	}
	if announced < 0 || entered < 0 {
		t.Fatalf("announced=%d entered=%d, want both present", announced, entered)
	}
	if announced > entered {
		t.Errorf("EventSpawned at %d is after EventETB at %d", announced, entered)
	}
}

// A refusal announces nothing. The event log is the table's memory,
// and a spawn line with no cards under it is a lie about what
// happened.
func TestRefusedSpawnAnnouncesNothing(t *testing.T) {
	g, alice, _ := spawnTestGame(t)
	before := len(g.Events)
	if _, err := g.SpawnCards(alice.ID, alice.ID, ZoneStack, devTemplate(), 1); err != ErrSpawnZoneUnsupported {
		t.Fatalf("err = %v, want ErrSpawnZoneUnsupported", err)
	}
	if _, err := g.SpawnCards(alice.ID, uuid.New(), ZoneHand, devTemplate(), 1); err != ErrPlayerNotFound {
		t.Fatalf("err = %v, want ErrPlayerNotFound", err)
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventSpawned {
			t.Fatalf("a refused spawn emitted %+v", ev)
		}
	}
}
