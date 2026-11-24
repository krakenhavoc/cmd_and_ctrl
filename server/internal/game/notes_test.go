package game

import (
	"slices"
	"testing"
)

// notes_test.go — #6097: the words an object's abilities note for it.

func TestNotesAreKeptOncePerWordAndOnlyOnTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	saga := pushBattlefieldForTest(g, me.ID, "Saga", "Enchantment — Saga", "")
	g.WithWriteLock(func() {
		if !g.AddNoteForEffect(saga, "Elf") || !g.AddNoteForEffect(saga, " Treefolk ") {
			t.Fatal("a fresh word was not noted")
		}
		if g.AddNoteForEffect(saga, "elf") || g.AddNoteForEffect(saga, "  ") {
			t.Error("a word already noted, or an empty one, was noted again")
		}
		if got := g.NotesForEffect(saga); !slices.Equal(got, []string{"Elf", "Treefolk"}) {
			t.Errorf("notes = %v, want [Elf Treefolk]", got)
		}
	})
	if !NotedIn([]string{"Elf"}, " ELF ") || NotedIn([]string{"Elf"}, "Elk") {
		t.Error("NotedIn compares wrongly")
	}

	// A copy taken now does not share the list with the live card.
	clone := g.Clone()
	g.WithWriteLock(func() { g.AddNoteForEffect(saga, "Ent") })
	clone.WithWriteLock(func() {
		if got := clone.NotesForEffect(saga); !slices.Equal(got, []string{"Elf", "Treefolk"}) {
			t.Errorf("clone notes = %v after a later note on the original", got)
		}
	})

	// Carried by the snapshot.
	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() {
		if got := restored.NotesForEffect(saga); !slices.Equal(got, []string{"Elf", "Treefolk", "Ent"}) {
			t.Errorf("restored notes = %v", got)
		}
	})

	// CR 400.7: the permanent that leaves keeps nothing, though what it
	// had is still readable as it last existed (CR 608.2h).
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, saga); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if got := g.NotesForEffect(saga); !slices.Equal(got, []string{"Elf", "Treefolk", "Ent"}) {
			t.Errorf("last-known notes = %v", got)
		}
		if g.AddNoteForEffect(saga, "Human") {
			t.Error("a card off the battlefield was given a note")
		}
	})
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == saga && len(c.Notes) != 0 {
			t.Errorf("the card in the graveyard still has notes %v", c.Notes)
		}
	}
}
