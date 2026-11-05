package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_dies_trigger_test.go — #4845. A face-down permanent has no
// text (CR 708.2), so it has none of its card's abilities while it is on
// the battlefield, and a leaves-the-battlefield ability looks back at the
// permanent as it last existed (CR 603.10a). A manifested Doomed Traveler
// that dies therefore has no dies trigger, even though the card is turned
// face up on its way to the graveyard (CR 708.9) and has its printed
// "When this creature dies" back there.
//
// Before #4845 the dying permanent's trigger identity restored only the
// card's oracle ID, and the catalog answered with the face-up card's
// triggers. The object snapshot a watcher's trigger carried also filled
// the face-down permanent's empty name from the card, naming the hidden
// card to the table.

const (
	hiddenDiesLabel   = "hidden card's own dies trigger"
	watcherDiesLabel  = "watcher: whenever a creature dies"
	faceDownWatcherID = "watcher-Watcher"
)

// diesTriggerCatalog installs the hidden card's printed "When this
// creature dies" (Doomed Traveler's shape) under hiddenKey, and a
// watcher's "Whenever a creature dies" under the watcher's key. Every
// other key has no triggers.
func diesTriggerCatalog(t *testing.T, hiddenKey string) {
	t.Helper()
	self := TriggeredAbility{
		Watches: []EventKind{EventLTB},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.CardID == source.InstanceID && ev.NewZone == ZoneGraveyard
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, hiddenDiesLabel)
		},
	}
	watcher := TriggeredAbility{
		Watches: []EventKind{EventLTB},
		// The fixtures kill only creatures, so "a creature" is not
		// re-checked here; lki is the WATCHER's own characteristics.
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.CardID != source.InstanceID && ev.NewZone == ZoneGraveyard
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			return NewTriggeredItem(source, watcherDiesLabel)
		},
	}
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		switch key {
		case hiddenKey:
			return []TriggeredAbility{self}
		case faceDownWatcherID:
			return []TriggeredAbility{watcher}
		}
		return nil
	})
}

// pendingTriggerLabelled is the queued trigger with this label from this
// source, or nil. With about set, only a trigger whose object snapshot
// names that object counts: a watcher fires once per creature in a wipe.
func pendingTriggerLabelled(g *Game, source uuid.UUID, label string, about uuid.UUID) *StackItem {
	var out *StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it.SourceCardID != source || it.Label != label {
				continue
			}
			if about != uuid.Nil && (it.Trigger == nil || it.Trigger.Object == nil || it.Trigger.Object.ID != about) {
				continue
			}
			out = it
		}
	})
	return out
}

// faceDownDiesFixture is one way the engine makes a face-down permanent:
// it returns the game, the permanent and the catalog key of the card it
// hides. FaceDownNone is the face-up control.
type faceDownDiesFixture struct {
	name string
	kind FaceDownKind
	make func(t *testing.T) (*Game, uuid.UUID, string)
}

// faceDownDiesFixtures covers every CR 708.2 permanent state the engine
// models (FaceDownKind.IsPermanentState), each made the way the engine
// makes it, plus megamorph (a morph with a counter on the way up), and
// a face-up control.
func faceDownDiesFixtures() []faceDownDiesFixture {
	const libOracle = "oracle-doomed-traveler"
	return []faceDownDiesFixture{
		{"face up (control)", FaceDownNone, func(t *testing.T) (*Game, uuid.UUID, string) {
			g := newActiveGame(t)
			id := turnableCreature(t, g, g.Seats[0], "Sheoldred")
			return g, id, "oracle-Sheoldred"
		}},
		{"manifest (CR 701.40a)", FaceDownManifested, func(t *testing.T) (*Game, uuid.UUID, string) {
			g, id := faceDownGame(t, "Doomed Traveler", "Creature — Human Soldier", libOracle)
			g.WithWriteLock(func() {
				if _, err := g.ManifestForEffect(g.Seats[0].ID); err != nil {
					t.Fatalf("manifest: %v", err)
				}
			})
			return g, id, libOracle
		}},
		{"cloak (CR 701.58a)", FaceDownCloaked, func(t *testing.T) (*Game, uuid.UUID, string) {
			g, id := faceDownGame(t, "Doomed Traveler", "Creature — Human Soldier", libOracle)
			g.WithWriteLock(func() {
				if _, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{id}, ZoneLibrary, ZoneEntryOptions{
					Controller: g.Seats[0].ID,
					FaceDown:   FaceDownCloaked,
				}); err != nil {
					t.Fatalf("cloak: %v", err)
				}
			})
			return g, id, libOracle
		}},
		{"morph cast face down (CR 702.37b)", FaceDownMorphed, func(t *testing.T) (*Game, uuid.UUID, string) {
			g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
			castFaceDownAndResolve(t, g, me, id)
			return g, id, morphOracle
		}},
		{"megamorph cast face down (CR 702.37b)", FaceDownMorphed, func(t *testing.T) (*Game, uuid.UUID, string) {
			g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, true))
			castFaceDownAndResolve(t, g, me, id)
			return g, id, morphOracle
		}},
		{"disguise cast face down (CR 702.168a)", FaceDownDisguised, func(t *testing.T) (*Game, uuid.UUID, string) {
			g, me, id := morphGame(t, morphOffer(FaceDownDisguised, morphCost, false))
			castFaceDownAndResolve(t, g, me, id)
			return g, id, morphOracle
		}},
		{"turned face down by an effect (CR 708.2a)", FaceDownTurned, func(t *testing.T) (*Game, uuid.UUID, string) {
			g := newActiveGame(t)
			id := turnableCreature(t, g, g.Seats[0], "Sheoldred")
			if got := g.TurnFaceDownForEffect(uuid.Nil, id); len(got) != 1 {
				t.Fatalf("turned %v face down, want 1", got)
			}
			return g, id, "oracle-Sheoldred"
		}},
	}
}

// requireDiesOutcome is the whole rule, read after `hidden` has died.
func requireDiesOutcome(t *testing.T, g *Game, hidden, watcher uuid.UUID, kind FaceDownKind, realName string) {
	t.Helper()
	if c, zone, _ := cardAnywhere(g, hidden); zone != ZoneGraveyard || c.FaceDown {
		t.Fatalf("the destroyed card is in %q face down=%v, want the graveyard face up (CR 708.9)", zone, c.FaceDown)
	}
	own := pendingTriggerLabelled(g, hidden, hiddenDiesLabel, uuid.Nil)
	if kind == FaceDownNone {
		if own == nil {
			t.Error("a face-up creature that died did not fire its own dies trigger")
		}
	} else if own != nil {
		t.Errorf("a %s permanent that died fired its card's printed dies trigger: "+
			"it had no text when it left (CR 708.2, CR 603.10a)", kind)
	}

	// Another permanent still sees a creature die: the face-down
	// permanent was a 2/2 creature. Its object snapshot names it as it
	// last existed, which for a face-down permanent is no name at all.
	seen := pendingTriggerLabelled(g, watcher, watcherDiesLabel, hidden)
	if seen == nil {
		t.Fatal("the watcher's \"whenever a creature dies\" did not fire for it, or carried no object snapshot")
	}
	wantName := realName
	if kind != FaceDownNone {
		wantName = ""
	}
	if got := seen.Trigger.Object.Name; got != wantName {
		t.Errorf("the dead creature's snapshot is named %q, want %q", got, wantName)
	}
}

// TestAFaceDownPermanentThatDiesHasNoDiesTrigger is #4845's repro, for
// every face-down kind and a face-up control, through a single destroy.
func TestAFaceDownPermanentThatDiesHasNoDiesTrigger(t *testing.T) {
	for _, f := range faceDownDiesFixtures() {
		t.Run(f.name, func(t *testing.T) {
			g, hidden, key := f.make(t)
			diesTriggerCatalog(t, key)
			watcher := watcherNamed(g, g.Seats[1].ID, "Watcher")
			c := battlefieldCardOf(t, g, hidden)
			if c.FaceDownKind != f.kind || c.FaceDownIsPermanent() != (f.kind != FaceDownNone) {
				t.Fatalf("the fixture is face down %q, want %q", c.FaceDownKind, f.kind)
			}
			realName := c.Name

			destroy(t, g, hidden)

			requireDiesOutcome(t, g, hidden, watcher, f.kind, realName)
		})
	}
}

// TestAFaceDownPermanentThatDiesInAWipeHasNoDiesTrigger is the same rule
// through a simultaneous exit (CR 603.10a's board wipe), whose harvest
// reads the batch's own copy of each permanent rather than the trigger
// identity record.
func TestAFaceDownPermanentThatDiesInAWipeHasNoDiesTrigger(t *testing.T) {
	for _, f := range faceDownDiesFixtures() {
		t.Run(f.name, func(t *testing.T) {
			g, hidden, key := f.make(t)
			diesTriggerCatalog(t, key)
			watcher := watcherNamed(g, g.Seats[1].ID, "Watcher")
			other := turnableCreature(t, g, g.Seats[1], "Griselbrand")
			realName := battlefieldCardOf(t, g, hidden).Name

			g.WithWriteLock(func() {
				if n := g.DestroyPermanentsForEffect([]uuid.UUID{hidden, other}); n != 2 {
					t.Fatalf("destroyed %d, want 2", n)
				}
			})

			requireDiesOutcome(t, g, hidden, watcher, f.kind, realName)
		})
	}
}

// TestAFaceDownTriggerIdentityRoundTripsTheSnapshot is the record
// itself: the face-down state rides the trigger identity through the
// snapshot (an additive field within the schema version, #522) and the
// undo clone, so a restored table still reads a dying permanent as it
// last existed.
func TestAFaceDownTriggerIdentityRoundTripsTheSnapshot(t *testing.T) {
	g := newActiveGame(t)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.lastKnownTriggerIdentity = map[uuid.UUID]triggerIdentityLKI{
			id: {OracleID: "oracle-doomed-traveler", FaceDown: FaceDownManifested},
		}
	})
	_, restored := roundTrip(t, g)
	if got := restored.lastKnownTriggerIdentity[id].FaceDown; got != FaceDownManifested {
		t.Errorf("the restored trigger identity is face down %q, want %q", got, FaceDownManifested)
	}
	if got := g.Clone().lastKnownTriggerIdentity[id].FaceDown; got != FaceDownManifested {
		t.Errorf("the cloned trigger identity is face down %q, want %q", got, FaceDownManifested)
	}
}
