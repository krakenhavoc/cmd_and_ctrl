package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// text_words_readers_test.go — #3872, the readers ADR 0160's 2026-10-15
// amendment listed as not reached and that now are: a creature type in a
// protection quality, a query naming its words, and a spell's own type
// line on the stack.

// The token names the creature type in the plural, so it is read and
// spelled back through the protection parser: Goblins → Elves, and a type
// with no regular plural keeps a spelling the parser reads as that type.
func TestTextWordsTranslateAProtectionFromACreatureType(t *testing.T) {
	var w *TextWords
	w = w.with(twChange(TextWordCreatureType, "Goblin", "Elf"))
	w = w.with(twChange(TextWordCreatureType, "Zombie", "Merfolk"))
	w = w.with(twChange(TextWordCreatureType, "Demon", "Dragon"))
	for in, want := range map[string]string{
		"protection from Goblins": "protection from Elves",
		"protection from Zombies": "protection from Merfolk",
		"protection from Demons":  "protection from Dragons",
		"protection from Humans":  "protection from Humans",
		"protection from black":   "protection from black",
	} {
		got := w.keyword(in)
		if got != want {
			t.Errorf("keyword(%q) = %q, want %q", in, got, want)
			continue
		}
		q, ok := ParseProtectionQuality(got)
		if !ok {
			t.Errorf("%q does not parse back as a protection", got)
			continue
		}
		if in != want && (q.Kind != ProtectionQualitySubtype || !slices.Contains([]string{"Elf", "Merfolk", "Dragon"}, q.Value)) {
			t.Errorf("%q parses as %+v, not the new creature type", got, q)
		}
	}
}

// A query names its words in the asking object's text, and is never
// written through.
func TestPermanentQueryReadsThroughText(t *testing.T) {
	src := &Card{}
	src.TextWords = (*TextWords)(nil).with(twChange(TextWordBasicLandType, "Island", "Forest")).
		with(twChange(TextWordColor, "U", "G"))
	q := PermanentQuery{Subtypes: []string{"Island"}, Colors: []string{"U"}, Types: []string{"land"}}
	got := q.ThroughTextOf(src)
	if !slices.Equal(got.Subtypes, []string{"Forest"}) || !slices.Equal(got.Colors, []string{"G"}) || !slices.Equal(got.Types, []string{"land"}) {
		t.Errorf("query through the text = %+v", got)
	}
	if !slices.Equal(q.Subtypes, []string{"Island"}) {
		t.Error("the query itself is not written through")
	}
	if back := q.ThroughTextOf(nil); !back.Equal(q) {
		t.Error("no reader leaves the query as it is")
	}
}

// A spell on the stack has no layer pass, so its changed type line and
// keywords are read off its baseline as it stands: a Goblin creature
// spell whose "Goblin" became "Elf" is an Elf spell before it resolves.
func TestASpellsChangedTypeLineIsReadOnTheStack(t *testing.T) {
	spell := Card{InstanceID: uuid.New(), Name: "Goblin Piker", TypeLine: "Creature — Goblin", Keywords: []string{"protection from black"}}
	if !spell.HasSubtype("Goblin") {
		t.Fatal("setup: the printed spell is a Goblin")
	}
	spell.TextWords = (*TextWords)(nil).with(twChange(TextWordCreatureType, "Goblin", "Elf")).
		with(twChange(TextWordColor, "B", "R"))
	if spell.HasSubtype("Goblin") || !spell.HasSubtype("Elf") {
		t.Errorf("the changed spell's subtypes = %v, want [Elf]", spell.Effective().Subtypes)
	}
	if !slices.Contains(spell.Effective().Abilities, "protection from red") {
		t.Errorf("the changed spell's keywords = %v", spell.Effective().Abilities)
	}
}
