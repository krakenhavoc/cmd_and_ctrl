package game

import (
	"slices"
	"testing"
)

// text_words_tokens_test.go — #4005, the engine half of ADR 0160's slice
// 4: a token description and a cast-permission filter read their words
// through the text whose words they are (CR 612.2). The cards are tested
// in cards/effects (text_word_tokens_test.go).

func twReader(changes ...TextWordChange) *Card {
	var w *TextWords
	for _, ch := range changes {
		w = w.with(ch)
	}
	return textReaderOf(w)
}

func TestTokenThroughTextReadsColoursTypesNameAndKeywords(t *testing.T) {
	reader := twReader(
		twChange(TextWordColor, "B", "U"),
		twChange(TextWordCreatureType, "Zombie", "Elf"),
		twChange(TextWordBasicLandType, "Island", "Swamp"),
	)
	zombie := Card{Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}}
	got := reader.TokenThroughText(zombie)
	if !slices.Equal(got.Colors, []string{"U"}) {
		t.Errorf("colours %v, want [U]", got.Colors)
	}
	if got.TypeLine != "Token Creature — Elf" || got.Name != "Elf" {
		t.Errorf("%q named %q, want a Token Creature — Elf named Elf (CR 111.4)", got.TypeLine, got.Name)
	}
	if zombie.Colors[0] != "B" || zombie.TypeLine != "Token Creature — Zombie" {
		t.Error("the template was written through")
	}

	// A token with a name of its own keeps it; a two-word type changes
	// the word it names; a landwalk keyword reads its land type.
	cadet := Card{Name: "Cadet", TypeLine: "Token Creature — Zombie Soldier", Keywords: []string{"islandwalk", "flying"}}
	got = reader.TokenThroughText(cadet)
	if got.Name != "Cadet" || got.TypeLine != "Token Creature — Elf Soldier" {
		t.Errorf("%q named %q, want a Token Creature — Elf Soldier named Cadet", got.TypeLine, got.Name)
	}
	if !slices.Equal(got.Keywords, []string{"swampwalk", "flying"}) {
		t.Errorf("keywords %v, want [swampwalk flying]", got.Keywords)
	}

	// Two printed colours that now read as one are one colour.
	both := reader.TokenThroughText(Card{Name: "Faerie", TypeLine: "Token Creature — Faerie", Colors: []string{"U", "B"}})
	if !slices.Equal(both.Colors, []string{"U"}) {
		t.Errorf("a blue and black token under black → blue has colours %v, want [U]", both.Colors)
	}

	// Printed words, no reader: the template as it is.
	var none *Card
	if got := none.TokenThroughText(zombie); got.Name != "Zombie" || got.Colors[0] != "B" {
		t.Errorf("a nil reader changed the token: %+v", got)
	}
	// A template with faces (a double-faced token) is not a description
	// this reads.
	dfc := Card{Name: "Incubator", TypeLine: "Token Artifact — Incubator", Faces: []Face{{Name: "Incubator"}, {Name: "Phyrexian Token", TypeLine: "Token Artifact Creature — Phyrexian"}}}
	if got := reader.TokenThroughText(dfc); got.TypeLine != dfc.TypeLine {
		t.Errorf("a double-faced template was read: %q", got.TypeLine)
	}
}

func TestPermissionFilterThroughTextReadsItsCreatureTypes(t *testing.T) {
	reader := twReader(twChange(TextWordCreatureType, "Zombie", "Elf"), twChange(TextWordCreatureType, "Merfolk", "Goblin"))
	f := PermissionFilter{CreatureOnly: true, CreatureType: "Zombie", CreatureTypesAny: [2]string{"Merfolk", "Druid"}}
	got := f.ThroughTextOf(reader)
	if got.CreatureType != "Elf" || got.CreatureTypesAny != [2]string{"Goblin", "Druid"} {
		t.Errorf("filter %+v, want Elf and [Goblin Druid]", got)
	}
	if f.CreatureType != "Zombie" {
		t.Error("the filter was written through")
	}
	if !got.Matches(Card{TypeLine: "Creature — Elf Druid"}) {
		t.Error("the read filter does not match an Elf Druid creature")
	}
	if got.Matches(Card{TypeLine: "Creature — Zombie Druid"}) {
		t.Error("the read filter still matches a Zombie")
	}
	// A chosen type (Realmwalker) is no word of the text.
	chosen := PermissionFilter{FromChosenType: true, CreatureType: "Zombie"}
	if got := chosen.ThroughTextOf(reader); got.CreatureType != "Zombie" {
		t.Errorf("a chosen type was translated to %q", got.CreatureType)
	}
	if got := f.ThroughTextOf(nil); got != f {
		t.Error("a nil reader changed the filter")
	}
}

func TestRelabelChangedWordsRespellsWholeWords(t *testing.T) {
	was := PermissionFilter{CreatureType: "Elf"}
	now := PermissionFilter{CreatureType: "Goblin"}
	got := relabelChangedWords("Cast an Elf spell, not an Elfhame or Elves (Elf Lord)", was, now)
	if want := "Cast a Goblin spell, not an Elfhame or Elves (Goblin Lord)"; got != want {
		t.Errorf("label %q, want %q", got, want)
	}
	got = relabelChangedWords("A Zombie, or cast a Zombie creature spell", PermissionFilter{CreatureType: "Zombie"}, PermissionFilter{CreatureType: "Elf"})
	if want := "An Elf, or cast an Elf creature spell"; got != want {
		t.Errorf("label %q, want %q", got, want)
	}
}
