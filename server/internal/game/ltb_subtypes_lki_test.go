package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ltb_subtypes_lki_test.go — #1679, CR 603.10a: a leaving permanent's
// subtypes as it last existed ride its EventLTB
// (Event.LastKnownSubtypes, Event.LastKnownAllCreatureTypes), so a
// tribal dies-trigger — Diregraf Captain's "whenever another Zombie
// you control dies" — counts a creature that was a Zombie only
// through a grant, after the move has left a plain Bear in the
// graveyard.
//
// Stamped by the same exitLKI as the card types (#1675) and the
// combat state (#1661); combatExitRoutes covers all three emit sites.

// grantForTest registers a battlefield-scoped layer effect on id and
// recomputes, failing the test if nothing registered.
func grantForTest(t *testing.T, g *Game, id uuid.UUID, label string, mods ...Mod) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id), mods,
			g.UntilEndOfTurnDuration(), label) {
			t.Fatalf("setup: %s registered nothing", label)
		}
		g.RecomputeLayersIfStaleLocked()
	})
}

// A creature that is a Zombie only through a type grant reports that
// it was one on every exit route; the card it leaves behind is not.
func TestLTBCarriesLastKnownSubtypesOnEveryExitRoute(t *testing.T) {
	for _, route := range combatExitRoutes {
		t.Run(route.name, func(t *testing.T) {
			g := newActiveGame(t)
			bear := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
			grantForTest(t, g, bear, "test — the bear is also a Zombie", AddSubtypesMod("Zombie"))

			seq := lastSeq(g)
			route.exit(t, g, bear)
			ev := ltbSince(t, g, seq, bear)
			if !slices.Contains(ev.LastKnownSubtypes, "Zombie") {
				t.Errorf("LastKnownSubtypes = %v, want Zombie", ev.LastKnownSubtypes)
			}
			if was, known := ev.WasSubtype("zombie"); !was || !known {
				t.Errorf("WasSubtype(zombie) = %v, %v; want true, true", was, known)
			}
			if was, known := ev.WasSubtype("Elf"); was || !known {
				t.Errorf("WasSubtype(Elf) = %v, %v; want false, true", was, known)
			}
			if c, ok := g.LookupCardForEffect(bear); ok && c.HasSubtype("Zombie") {
				t.Error("the card off the battlefield is still a Zombie; the event is only needed because it is not")
			}
		})
	}
}

// A creature granted every creature type (Maskwood Nexus' shape)
// reports every creature type through the flag — and only creature
// types: it was not a Forest or an Equipment.
func TestLTBLastKnownAllCreatureTypesFromAGrant(t *testing.T) {
	g := newActiveGame(t)
	bear := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	grantForTest(t, g, bear, "test — the bear is every creature type", AllCreatureTypesMod())

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	ev := ltbSince(t, g, seq, bear)
	if !ev.LastKnownAllCreatureTypes {
		t.Fatal("LastKnownAllCreatureTypes = false for a creature granted every creature type")
	}
	for _, tribe := range []string{"Zombie", "Goblin", "elf", "Egg"} {
		if was, _ := ev.WasSubtype(tribe); !was {
			t.Errorf("WasSubtype(%s) = false for an every-type creature", tribe)
		}
	}
	for _, notTribe := range []string{"Forest", "Equipment"} {
		if was, _ := ev.WasSubtype(notTribe); was {
			t.Errorf("WasSubtype(%s) = true; every creature type is not every subtype", notTribe)
		}
	}
	if c, _ := g.LookupCardForEffect(bear); c.HasSubtype("Zombie") {
		t.Fatal("setup: the graveyard card should be a plain Bear")
	}
}

// A printed changeling (CR 702.73a) was every creature type, and the
// flag says so without the ~345 types being written anywhere.
func TestLTBLastKnownSubtypesOfAPrintedChangeling(t *testing.T) {
	g := newActiveGame(t)
	shifter := pushCombatant(t, g, g.Seats[0], "Changeling Outcast", 1, 1, KeywordChangeling)

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(shifter); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	ev := ltbSince(t, g, seq, shifter)
	if !ev.LastKnownAllCreatureTypes {
		t.Error("LastKnownAllCreatureTypes = false for a printed changeling")
	}
	if len(ev.LastKnownSubtypes) > 2 {
		t.Errorf("LastKnownSubtypes = %v: the changeling's types ride the flag, not the list", ev.LastKnownSubtypes)
	}
	if was, _ := ev.WasSubtype("Vampire"); !was {
		t.Error("WasSubtype(Vampire) = false for a printed changeling")
	}
}

// WasSubtype answers "unknown" for anything that is not an EventLTB
// and for an EventLTB with no last-known stamp; a stamped event with
// no subtypes is a known "no".
func TestWasSubtypeIsUnknownWithoutAStamp(t *testing.T) {
	if _, known := (Event{Kind: EventLTB}).WasSubtype("Zombie"); known {
		t.Error("an unstamped EventLTB claimed to know its subtypes")
	}
	if _, known := (Event{Kind: EventETB, LastKnownTypes: []string{"Creature"}, LastKnownSubtypes: []string{"Zombie"}}).WasSubtype("Zombie"); known {
		t.Error("a non-LTB event claimed last-known subtypes")
	}
	if was, known := (Event{Kind: EventLTB, LastKnownTypes: []string{"Creature"}}).WasSubtype("Zombie"); was || !known {
		t.Errorf("a stamped creature with no subtypes: WasSubtype = %v, %v; want false, true", was, known)
	}
}

// The subtypes live on the event and nowhere else, so the event log
// carries them through a persisted snapshot, an undo clone, and the
// undo and redo themselves — and a trigger context's copy of the
// event does not share the slice with the log.
func TestLTBLastKnownSubtypesSurviveSnapshotCloneAndUndo(t *testing.T) {
	g := newActiveGame(t)
	bear := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	grantForTest(t, g, bear, "test — the bear is also a Zombie", AddSubtypesMod("Zombie"))
	shifter := pushCombatant(t, g, g.Seats[0], "Changeling Outcast", 1, 1, KeywordChangeling)
	before := g.Clone()
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
		if err := g.DestroyPermanentForEffect(shifter); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	want := ltbSince(t, g, seq, bear).LastKnownSubtypes
	if !slices.Contains(want, "Zombie") {
		t.Fatalf("LastKnownSubtypes = %v, want the granted Zombie", want)
	}
	if !ltbSince(t, g, seq, shifter).LastKnownAllCreatureTypes {
		t.Fatal("the changeling's LTB lost its every-type flag")
	}

	check := func(label string, gg *Game) {
		t.Helper()
		if got := ltbSince(t, gg, seq, bear).LastKnownSubtypes; !slices.Equal(got, want) {
			t.Errorf("%s: LastKnownSubtypes = %v, want %v", label, got, want)
		}
		if !ltbSince(t, gg, seq, shifter).LastKnownAllCreatureTypes {
			t.Errorf("%s: LastKnownAllCreatureTypes lost", label)
		}
	}
	_, restored := roundTrip(t, g)
	check("snapshot round trip", restored)
	after := g.Clone()
	check("clone", after)

	g.RestoreFrom(before)
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventLTB && (ev.CardID == bear || ev.CardID == shifter) {
			t.Fatalf("undo kept the EventLTB: %+v", ev)
		}
	}
	g.RestoreFrom(after)
	check("restore of the post-destroy clone", g)

	tc := &TriggerContext{Event: ltbSince(t, g, seq, bear)}
	cp := cloneTriggerContext(tc)
	cp.Event.LastKnownSubtypes[0] = "Mutated"
	if tc.Event.LastKnownSubtypes[0] == "Mutated" {
		t.Error("cloneTriggerContext shares LastKnownSubtypes with the original")
	}
}
