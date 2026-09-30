package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ltb_types_lki_test.go — #1675, CR 603.10a: a leaving permanent's
// card types as it last existed ride its EventLTB
// (Event.LastKnownTypes), so "whenever a creature dies" can count a
// permanent that was a creature only because of an effect — a crewed
// Vehicle, an animated manland — after the move has left it an
// artifact or a land in the graveyard.
//
// Stamped by the same exitLKI the combat state rides (#1661), on all
// three EventLTB emit sites; combatExitRoutes (combat_lki_test.go)
// covers every one of them.

// pushAnimatedLand puts a land on the battlefield and makes it a 3/3
// creature that is still a land until end of turn — Mutavault's
// shape, through the same scoped-effect record a card would use.
func pushAnimatedLand(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	land := NewCard("Test Manland", owner.ID)
	land.TypeLine = "Land"
	g.Battlefield.PushTop(land)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(land.InstanceID),
			[]Mod{AddTypesMod("Creature"), SetBasePowerMod(3), SetBaseToughnessMod(3)},
			g.UntilEndOfTurnDuration(), "test — the land becomes a 3/3 creature") {
			t.Fatal("setup: the animation registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c, _ := g.battlefieldCardLocked(land.InstanceID); !c.IsCreature() || !c.IsLand() {
		t.Fatal("setup: the animated land is not a land creature")
	}
	return land.InstanceID
}

// An animated land leaving by any route reports that it was a
// creature and a land; the card it leaves behind is only a land.
func TestLTBCarriesLastKnownTypesOnEveryExitRoute(t *testing.T) {
	for _, route := range combatExitRoutes {
		t.Run(route.name, func(t *testing.T) {
			g := newActiveGame(t)
			land := pushAnimatedLand(t, g, g.Seats[0])

			seq := lastSeq(g)
			route.exit(t, g, land)
			ev := ltbSince(t, g, seq, land)
			if !slices.Contains(ev.LastKnownTypes, "Creature") || !slices.Contains(ev.LastKnownTypes, "Land") {
				t.Errorf("LastKnownTypes = %v, want Creature and Land", ev.LastKnownTypes)
			}
			if was, known := ev.WasType("creature"); !was || !known {
				t.Errorf("WasType(creature) = %v, %v; want true, true", was, known)
			}
			if c, ok := g.LookupCardForEffect(land); ok && c.IsCreature() {
				t.Error("the card off the battlefield is still a creature; the event is only needed because it is not")
			}
		})
	}
}

// The other direction: a printed creature that an effect made a
// non-creature before it left reports what it was then, not what it
// prints.
func TestLTBLastKnownTypesSeeACreatureThatStoppedBeingOne(t *testing.T) {
	g := newActiveGame(t)
	bear := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
			[]Mod{RemoveTypesMod("Creature"), AddTypesMod("Artifact")},
			g.UntilEndOfTurnDuration(), "test — the bear becomes a noncreature artifact") {
			t.Fatal("setup: the effect registered nothing")
		}
		g.RecomputeLayersIfStaleLocked()
	})

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	ev := ltbSince(t, g, seq, bear)
	if was, known := ev.WasType("creature"); was || !known {
		t.Errorf("WasType(creature) = %v, %v; want false, true — it was not a creature when it left", was, known)
	}
	if was, _ := ev.WasType("artifact"); !was {
		t.Errorf("WasType(artifact) = false, want true (LastKnownTypes %v)", ev.LastKnownTypes)
	}
	if c, _ := g.LookupCardForEffect(bear); !c.IsCreature() {
		t.Fatal("setup: the graveyard card should read as the creature it prints")
	}
}

// WasType answers "unknown" for anything that is not an EventLTB, and
// for an EventLTB that carries no types (logged before the field
// existed): the caller falls back rather than reading a false "no".
func TestWasTypeIsUnknownWithoutAStamp(t *testing.T) {
	if _, known := (Event{Kind: EventLTB}).WasType("creature"); known {
		t.Error("an unstamped EventLTB claimed to know its types")
	}
	if _, known := (Event{Kind: EventETB, LastKnownTypes: []string{"Creature"}}).WasType("creature"); known {
		t.Error("a non-LTB event claimed last-known types")
	}
}

// The types live on the event and nowhere else, so the event log is
// what carries them through a persisted snapshot, an undo clone and
// the undo itself.
func TestLTBLastKnownTypesSurviveSnapshotCloneAndUndo(t *testing.T) {
	g := newActiveGame(t)
	land := pushAnimatedLand(t, g, g.Seats[0])
	before := g.Clone()
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(land); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	want := ltbSince(t, g, seq, land).LastKnownTypes
	if !slices.Contains(want, "Creature") {
		t.Fatalf("LastKnownTypes = %v, want the animated land's Creature", want)
	}

	_, restored := roundTrip(t, g)
	if got := ltbSince(t, restored, seq, land).LastKnownTypes; !slices.Equal(got, want) {
		t.Errorf("snapshot round trip: LastKnownTypes = %v, want %v", got, want)
	}
	after := g.Clone()
	if got := ltbSince(t, after, seq, land).LastKnownTypes; !slices.Equal(got, want) {
		t.Errorf("clone: LastKnownTypes = %v, want %v", got, want)
	}

	// Undo the destroy: the event is gone with it, and the land is
	// back as the animated creature it was.
	g.RestoreFrom(before)
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventLTB && ev.CardID == land {
			t.Fatalf("undo kept the EventLTB: %+v", ev)
		}
	}
	// Redo by restoring the later clone: the stamp comes back intact.
	g.RestoreFrom(after)
	if got := ltbSince(t, g, seq, land).LastKnownTypes; !slices.Equal(got, want) {
		t.Errorf("restore of the post-destroy clone: LastKnownTypes = %v, want %v", got, want)
	}
}
