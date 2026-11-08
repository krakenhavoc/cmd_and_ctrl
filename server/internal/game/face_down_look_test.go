package game

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// face_down_look_test.go — #5251: "you may look at face-down creatures
// you don't control any time", the engine's half. The wire's half (who
// actually gets the card) is protocol/face_down_look_view_test.go.

const lookStaticOracle = "oracle-found-footage-test"

// stubLookCatalog makes lookStaticOracle a permanent with the static.
func stubLookCatalog(t *testing.T) {
	t.Helper()
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if key == lookStaticOracle {
			return &CardDef{LooksAtFaceDownCreatures: true}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogLookup = prev })
}

// pushLookTestPermanent puts a face-up permanent onto the battlefield
// under `controller`, known to every seat as a public permanent is.
func pushLookTestPermanent(g *Game, controller uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	c := NewCard(name, controller)
	c.Controller = controller
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.Power, c.Toughness = 4, 5
	for _, p := range g.Seats {
		c.AddKnower(p.ID)
	}
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

// pushLookTestFaceDown puts a creature under `controller` and turns it
// face down through the engine's own door, so KnownBy is what a real
// face-down permanent's is: its controller alone.
func pushLookTestFaceDown(t *testing.T, g *Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := pushLookTestPermanent(g, controller, name, "Creature — Phyrexian Praetor", "")
	var turned []uuid.UUID
	g.WithWriteLock(func() { turned = g.TurnFaceDownForEffect(uuid.Nil, id) })
	if len(turned) != 1 {
		t.Fatalf("%s was not turned face down", name)
	}
	return id
}

func lookGrants(g *Game) map[uuid.UUID][]uuid.UUID {
	var out map[uuid.UUID][]uuid.UUID
	g.ReadSnapshot(func() { out = g.FaceDownLookGrantsLocked() })
	return out
}

// The static (Found Footage): its controller, and only its controller,
// may look at every face-down creature another player controls, while
// the source is on the battlefield. Nothing is written into KnownBy.
func TestALookStaticCoversFaceDownCreaturesOthersControl(t *testing.T) {
	stubLookCatalog(t)
	g := newFourPlayerActiveGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	src := pushLookTestPermanent(g, me.ID, "Found Footage", "Artifact — Clue", lookStaticOracle)
	theirs := pushLookTestFaceDown(t, g, opp.ID, "Sheoldred, the Apocalypse")
	mine := pushLookTestFaceDown(t, g, me.ID, "Griselbrand")
	faceUp := pushLookTestPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "")

	got := lookGrants(g)
	if want := []uuid.UUID{me.ID}; !slices.Equal(got[theirs], want) {
		t.Errorf("an opponent's face-down creature: lookers %v, want %v", got[theirs], want)
	}
	if len(got[mine]) != 0 {
		t.Errorf("my own face-down creature: lookers %v, want none (CR 708.5 already lets me)", got[mine])
	}
	if len(got[faceUp]) != 0 {
		t.Errorf("a face-up creature: lookers %v, want none", got[faceUp])
	}
	g.ReadSnapshot(func() {
		if !g.MayLookAtFaceDownCreatureLocked(me.ID, theirs) {
			t.Error("the grant's holder may not look at the opponent's face-down creature")
		}
		if g.MayLookAtFaceDownCreatureLocked(other.ID, theirs) {
			t.Error("another opponent may look at it through someone else's grant")
		}
	})
	c, _ := g.LookupCardForEffect(theirs)
	if c.IsKnownTo(me.ID) || c.IsKnownTo(other.ID) || !c.IsKnownTo(opp.ID) {
		t.Errorf("KnownBy changed: me %v other %v controller %v, want only the controller",
			c.IsKnownTo(me.ID), c.IsKnownTo(other.ID), c.IsKnownTo(opp.ID))
	}

	// The source leaves: the look ends with it, at once.
	g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(src) })
	if got := lookGrants(g); len(got) != 0 {
		t.Errorf("after the source left: grants %v, want none", got)
	}
}

// A face-down permanent that is not a creature (Yedora's Forest, a
// listed land) is not a "face-down creature", and a face-down creature
// turned face up is simply public again.
func TestALookCoversOnlyFaceDownCreatures(t *testing.T) {
	stubLookCatalog(t)
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLookTestPermanent(g, me.ID, "Found Footage", "Artifact — Clue", lookStaticOracle)
	forest := pushLookTestFaceDown(t, g, opp.ID, "Yedora's Forest")
	morph := pushLookTestFaceDown(t, g, opp.ID, "Sheoldred, the Apocalypse")
	g.WithWriteLock(func() {
		idx := findCardOnBattlefield(g, forest)
		c := &g.Battlefield.Cards[idx]
		c.SetFaceDownListed(FaceDownTurned, &FaceDownListing{Types: []string{"Land"}, Subtypes: []string{"Forest"}})
		c.effective = nil
	})
	if got := lookGrants(g); len(got[forest]) != 0 {
		t.Errorf("a face-down land: lookers %v, want none", got[forest])
	}
	if got := lookGrants(g); len(got[morph]) != 1 {
		t.Fatalf("the face-down creature: lookers %v, want one", got[morph])
	}
	g.WithWriteLock(func() {
		if !g.TurnFaceUpForEffect(uuid.Nil, opp.ID, morph) {
			t.Fatal("could not turn the creature face up")
		}
	})
	if got := lookGrants(g); len(got[morph]) != 0 {
		t.Errorf("a creature turned face up: lookers %v, want none", got[morph])
	}
}

// The until-end-of-turn grant (Lumbering Laundry): a record naming the
// activator, read like the static, gone after cleanup, and kept by a
// restore point taken while it lasts.
func TestALookGrantLastsTheTurnAndSurvivesARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pushLookTestFaceDown(t, g, opp.ID, "Sheoldred, the Apocalypse")
	mine := pushLookTestFaceDown(t, g, me.ID, "Griselbrand")
	g.WithWriteLock(func() {
		if !g.LookAtFaceDownCreaturesThisTurnForEffect(uuid.Nil, me.ID, "Lumbering Laundry") {
			t.Fatal("no record was written")
		}
	})
	got := lookGrants(g)
	if want := []uuid.UUID{me.ID}; !slices.Equal(got[theirs], want) {
		t.Errorf("an opponent's face-down creature: lookers %v, want %v", got[theirs], want)
	}
	if len(got[mine]) != 0 {
		t.Errorf("my own face-down creature: lookers %v, want none", got[mine])
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game holding a look grant is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if got := lookGrants(restored); !slices.Equal(got[theirs], []uuid.UUID{me.ID}) {
		t.Errorf("restored: lookers %v, want the activator", got[theirs])
	}
	var other2 bool
	restored.ReadSnapshot(func() { other2 = restored.MayLookAtFaceDownCreatureLocked(other.ID, theirs) })
	if other2 {
		t.Error("restored: another opponent may look")
	}

	// CR 514.2: cleanup ends it.
	g.WithWriteLock(func() { g.sweepScopedEffectsLocked(true) })
	if got := lookGrants(g); len(got) != 0 {
		t.Errorf("after cleanup: grants %v, want none", got)
	}
}

func TestALookGrantRefusesAPlayerWhoIsNotSeated(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.LookAtFaceDownCreaturesThisTurnForEffect(uuid.Nil, uuid.Nil, "x") {
			t.Error("a record with no player was written")
		}
		if g.LookAtFaceDownCreaturesThisTurnForEffect(uuid.Nil, uuid.New(), "x") {
			t.Error("a record for a stranger was written")
		}
	})
}
