package game

import (
	"testing"

	"github.com/google/uuid"
)

// TestSpellCardRestrictionNamesOneCardByInstance: Ice Cauldron's "spend this
// mana only to cast the last card exiled with this artifact" is the cast of
// that card, not of another card with the same name, and never an ability.
func TestSpellCardRestrictionNamesOneCardByInstance(t *testing.T) {
	exiled := Card{InstanceID: uuid.New(), Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}"}
	twin := Card{InstanceID: uuid.New(), Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}"}
	tok := ManaToken{Color: "R", Restrictions: []string{ManaRestrictSpellCard(exiled.InstanceID)}}
	if !ManaSpendForCast(exiled).allowsToken(tok) {
		t.Error("the named card cannot be paid for with the mana")
	}
	if ManaSpendForCast(twin).allowsToken(tok) {
		t.Error("a second card with the same name was paid for with the mana")
	}
	if ManaSpendForAbility(exiled).allowsToken(tok) {
		t.Error("an activation was paid for with mana for a cast")
	}
	if ManaSpendForCast(exiled).allowsToken(ManaToken{Color: "R", Restrictions: []string{ManaRestrictSpellCard(uuid.Nil)}}) {
		t.Error("the nil card matched a real one")
	}
}

// TestPrefixedNoteReplacesAndTakesBack: a prefixed note is a single value a
// later write replaces, and an empty value takes it back, without touching
// the other notes.
func TestPrefixedNoteReplacesAndTakesBack(t *testing.T) {
	g := newActiveGame(t)
	id := pushBattlefieldForTest(g, g.Seats[0].ID, "Thing", "Artifact", "")
	g.WithWriteLock(func() {
		g.AddNoteForEffect(id, "other")
		g.SetPrefixedNoteForEffect(id, "k:", "one")
		g.SetPrefixedNoteForEffect(id, "k:", "two")
		if v, ok := g.PrefixedNoteForEffect(id, "k:"); !ok || v != "two" {
			t.Fatalf("note = %q, %v; want two", v, ok)
		}
		if n := g.NotesForEffect(id); len(n) != 2 {
			t.Errorf("notes = %v, want the other note and one k: note", n)
		}
		g.SetPrefixedNoteForEffect(id, "k:", "")
		if _, ok := g.PrefixedNoteForEffect(id, "k:"); ok {
			t.Error("the note was not taken back")
		}
		if n := g.NotesForEffect(id); len(n) != 1 || n[0] != "other" {
			t.Errorf("notes = %v, want [other]", n)
		}
	})
}
