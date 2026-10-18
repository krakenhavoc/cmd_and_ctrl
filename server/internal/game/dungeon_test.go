package game

import (
	"testing"

	"github.com/google/uuid"
)

// dungeon_test.go pins the engine half of ADR 0203 that the catalog's
// one dungeon cannot reach: choosing among several dungeons (CR 309.2a)
// and RegisterDungeon's guards. The room abilities and the Undercity are
// pinned in cards/effects/initiative_*_test.go.

// Two plain dungeons, registered for this test binary only (their keys
// are in the test namespace). Nothing else in the package ventures
// without naming a dungeon, so no other test is offered them.
var (
	testDungeonA = func() string {
		key := testEffectKeyPrefix + "dungeon-a"
		RegisterDungeon(DungeonDef{Key: key, Name: "Dungeon A", Rooms: []DungeonRoom{
			{Key: "a1", Name: "A One", Text: "Nothing.", LeadsTo: []string{"a2"}},
			{Key: "a2", Name: "A Two", Text: "Nothing."},
		}})
		return key
	}()
	testDungeonB = func() string {
		key := testEffectKeyPrefix + "dungeon-b"
		RegisterDungeon(DungeonDef{Key: key, Name: "Dungeon B", Rooms: []DungeonRoom{
			{Key: "b1", Name: "B One", Text: "Nothing.", LeadsTo: []string{"b2"}},
			{Key: "b2", Name: "B Two", Text: "Nothing."},
		}})
		return key
	}()
)

// TestAPlainVentureAsksWhichDungeon: with more than one dungeon a plain
// venture may enter, the player chooses one (CR 309.2a), and the answer
// puts the marker on that dungeon's topmost room. A dungeon marked
// venture-into-only is not offered.
func TestAPlainVentureAsksWhichDungeon(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.VentureForEffect(me.ID, uuid.Nil, ""); err != nil {
			t.Fatalf("VentureForEffect: %v", err)
		}
	})
	var c *PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == PendingChoiceOptionPick && pc.Chooser == me.ID {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("no dungeon prompt")
	}
	idx := -1
	for i, o := range c.PickOptions {
		if o.Dungeon == carriedTestDungeon {
			t.Errorf("a venture-into-only dungeon was offered to a plain venture")
		}
		if o.Dungeon == testDungeonB {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("Dungeon B not offered: %+v", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if me.Dungeon == nil || me.Dungeon.Key != testDungeonB || me.Dungeon.Room != "b1" {
		t.Fatalf("dungeon = %+v, want Dungeon B at b1", me.Dungeon)
	}
	_ = testDungeonA
}

// TestVentureIntoAQualityEntersOnlyThatDungeon is CR 701.49d: "venture
// into Undercity" enters the named dungeon with no question.
func TestVentureIntoAQualityEntersOnlyThatDungeon(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.VentureForEffect(me.ID, uuid.Nil, carriedTestDungeon); err != nil {
			t.Fatalf("VentureForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a venture into a named dungeon asked a question")
	}
	if me.Dungeon == nil || me.Dungeon.Key != carriedTestDungeon || me.Dungeon.Room != "first" {
		t.Fatalf("dungeon = %+v, want the named dungeon at its first room", me.Dungeon)
	}
}

func TestRegisterDungeonRefusesAMalformedDungeon(t *testing.T) {
	for name, d := range map[string]DungeonDef{
		"no rooms":         {Key: testEffectKeyPrefix + "bad-1", Name: "Bad"},
		"unknown arrow":    {Key: testEffectKeyPrefix + "bad-2", Name: "Bad", Rooms: []DungeonRoom{{Key: "x", Name: "X", LeadsTo: []string{"y"}}}},
		"two last rooms":   {Key: testEffectKeyPrefix + "bad-3", Name: "Bad", Rooms: []DungeonRoom{{Key: "x", Name: "X"}, {Key: "y", Name: "Y"}}},
		"unreachable room": {Key: testEffectKeyPrefix + "bad-4", Name: "Bad", Rooms: []DungeonRoom{{Key: "x", Name: "X"}, {Key: "y", Name: "Y", LeadsTo: []string{"x"}}}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("RegisterDungeon accepted %s", name)
				}
			}()
			RegisterDungeon(d)
		})
	}
}
