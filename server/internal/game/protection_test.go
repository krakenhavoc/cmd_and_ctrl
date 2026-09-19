package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// protection_test.go — #662, the READER. CR 702.16's four checks all
// go through one closed grammar, so the grammar gets its own file:
// everything else in the protection suite is about what the engine
// DOES with a quality, and this is about what a quality IS.
//
// The negative cases are the important half. A quality the grammar
// cannot parse must mint NO token, because the same parser answers
// the ADR 0037 coverage signal — a card whose protection the engine
// does not enforce has to keep saying so.

func TestProtectionGrammarParsesTheClosedSet(t *testing.T) {
	cases := []struct {
		token string
		kind  ProtectionQualityKind
		value string
	}{
		{"protection from red", ProtectionQualityColor, "R"},
		{"protection from white", ProtectionQualityColor, "W"},
		{"protection from blue", ProtectionQualityColor, "U"},
		{"protection from black", ProtectionQualityColor, "B"},
		{"protection from green", ProtectionQualityColor, "G"},
		{"protection from artifacts", ProtectionQualityCardType, "artifact"},
		{"protection from artifact", ProtectionQualityCardType, "artifact"},
		{"protection from creatures", ProtectionQualityCardType, "creature"},
		{"protection from sorceries", ProtectionQualityCardType, "sorcery"},
		{"protection from Demons", ProtectionQualitySubtype, "Demon"},
		{"protection from Dragons", ProtectionQualitySubtype, "Dragon"},
		{"protection from Faeries", ProtectionQualitySubtype, "Faerie"},
		{"protection from Elves", ProtectionQualitySubtype, "Elf"},
		{"protection from Goblin", ProtectionQualitySubtype, "Goblin"},
		{"protection from everything", ProtectionQualityEverything, ""},
		// Case-insensitive on the fixed head, so a card file and an
		// oracle line agree.
		{"Protection from Red", ProtectionQualityColor, "R"},
	}
	for _, tc := range cases {
		q, ok := ParseProtectionQuality(tc.token)
		if !ok {
			t.Errorf("%q did not parse", tc.token)
			continue
		}
		if q.Kind != tc.kind || q.Value != tc.value {
			t.Errorf("%q = (%v, %q), want (%v, %q)", tc.token, q.Kind, q.Value, tc.kind, tc.value)
		}
	}
}

// TestProtectionGrammarRefusesWhatItCannotEnforce is the half that
// keeps ADR 0037 honest. Each of these is a real printed quality; the
// engine has no way to test any of them against a source, so it must
// mint no token and leave the card flagged.
func TestProtectionGrammarRefusesWhatItCannotEnforce(t *testing.T) {
	for _, token := range []string{
		"protection",                        // no quality at all
		"protection from",                   // ditto
		"protection from monocolored",       // Sphinx of the Guildpact
		"protection from multicolored",      // Ghostly Prison-class wording
		"protection from all colors",        // Progenitus's older wording
		"protection from opponents",         // a player quality, #929
		"protection from the chosen player", // True-Name Nemesis, #929
		"protection from Zubera the Ascended",
		"flying",
		"",
		"Target creature gains protection from the color of your choice until end of turn.",
	} {
		if q, ok := ParseProtectionQuality(token); ok {
			t.Errorf("%q parsed as %+v; the grammar must refuse it so the card stays flagged", token, q)
		}
	}
}

// TestProtectionTokensSplitsOneClauseIntoTwoAbilities is CR 702.16m
// through Baneslayer Angel's printed line: "protection from Demons
// and from Dragons" is two abilities, and a Demon Dragon is refused
// by either.
func TestProtectionTokensSplitsOneClauseIntoTwoAbilities(t *testing.T) {
	got, ok := ProtectionTokens("Protection from Demons and from Dragons")
	if !ok {
		t.Fatal("the printed clause did not parse")
	}
	want := []string{"protection from Demons", "protection from Dragons"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	// Mirran Crusader.
	if got, ok := ProtectionTokens("protection from black and from green"); !ok ||
		!reflect.DeepEqual(got, []string{"protection from black", "protection from green"}) {
		t.Errorf("two colours: got %q (%v)", got, ok)
	}
}

// A clause is all-or-nothing. Half a protection is not a weaker
// protection, it is a different card, and shipping the half the
// grammar happens to understand would be the one direction this repo
// never errs in.
func TestProtectionTokensRefusesAPartlyUnparseableClause(t *testing.T) {
	if got, ok := ProtectionTokens("protection from red and from monocolored"); ok {
		t.Errorf("got %q, want the whole clause refused", got)
	}
}

func TestCanonicalKeywordsRefusesBareProtection(t *testing.T) {
	if kw, ok := CanonicalKeyword("Protection"); ok {
		t.Errorf(`CanonicalKeyword("Protection") = (%q, true); a quality-less badge is a promise the rules layer cannot keep`, kw)
	}
	if kws, ok := CanonicalKeywords("protection from Demons and from Dragons"); !ok || len(kws) != 2 {
		t.Errorf("CanonicalKeywords on a two-quality clause = (%q, %v)", kws, ok)
	}
	// The singular form must not pick one of two.
	if kw, ok := CanonicalKeyword("protection from Demons and from Dragons"); ok {
		t.Errorf("CanonicalKeyword picked %q out of a two-ability clause", kw)
	}
}

func TestProtectionFromColorMintsTheToken(t *testing.T) {
	if got := ProtectionFromColor("R"); got != "protection from red" {
		t.Errorf(`ProtectionFromColor("R") = %q`, got)
	}
	if got := ProtectionFromColor("C"); got != "" {
		t.Errorf(`ProtectionFromColor("C") = %q, want ""`, got)
	}
}

// TestProtectionQualitiesReadsGrantsAndPrintedAlike: the reader is the
// same ability list HasKeyword walks, so an Equipment's layer-6 grant
// and a printed keyword answer the same way. A grant that only the
// printed road could see is how The Wandering Rescuer shipped inert.
func TestProtectionQualitiesReadsGrantsAndPrintedAlike(t *testing.T) {
	c := NewCard("Bear", uuid.New())
	c.TypeLine = "Creature — Bear"
	c.Keywords = []string{"flying", "protection from red", "protection from monocolored"}

	qs := ProtectionQualities(&c)
	if len(qs) != 1 || qs[0].Value != "R" {
		t.Fatalf("printed road: got %+v, want just protection from red", qs)
	}
	if !HasProtection(&c) {
		t.Error("HasProtection must agree with the reader")
	}

	// The layered road: whatever the layer engine resolved.
	c.effective = &Characteristic{Abilities: []string{"protection from Demons"}}
	qs = ProtectionQualities(&c)
	if len(qs) != 1 || qs[0].Value != "Demon" {
		t.Fatalf("layered road: got %+v", qs)
	}
}

// TestProtectedFromMatchesTheSourcesCharacteristics is the predicate
// all four DEBT checks call, exercised on its own so a failure
// upstream is never ambiguous about which half broke.
func TestProtectedFromMatchesTheSourcesCharacteristics(t *testing.T) {
	proRed := NewCard("Pro Red", uuid.New())
	proRed.TypeLine = "Creature — Bear"
	proRed.Keywords = []string{"protection from red"}

	red := &Characteristic{Colors: []string{"R"}, Types: []string{"Instant"}}
	white := &Characteristic{Colors: []string{"W"}, Types: []string{"Instant"}}
	colourless := &Characteristic{Types: []string{"Artifact"}}

	if !ProtectedFrom(&proRed, red) {
		t.Error("a red source must be refused")
	}
	if ProtectedFrom(&proRed, white) {
		t.Error("a white source must not be")
	}
	if ProtectedFrom(&proRed, colourless) {
		t.Error("a colourless source must not be")
	}
	// An unknown source (nil LKI) prevents nothing — errs weaker.
	if ProtectedFrom(&proRed, nil) {
		t.Error("an unknown source must not match")
	}

	everything := NewCard("Progenitus", uuid.New())
	everything.TypeLine = "Creature — Hydra Avatar"
	everything.Keywords = []string{"protection from everything"}
	for _, src := range []*Characteristic{red, white, colourless, nil} {
		if !ProtectedFrom(&everything, src) {
			t.Errorf("protection from everything must match %+v, including an unknown source (CR 702.16j)", src)
		}
	}
}

// TestChangelingCountsAsTheSubtype is CR 702.73a meeting CR 702.16:
// Baneslayer Angel's quality is a SUBTYPE, and #939 put "is every
// creature type" in a layer-4 flag rather than in ~345 subtypes — so
// the matcher has to ask the flag as well as the list, or a Mistform
// Ultimus walks straight past her.
func TestChangelingCountsAsTheSubtype(t *testing.T) {
	angel := NewCard("Baneslayer Angel", uuid.New())
	angel.TypeLine = "Creature — Angel"
	angel.Keywords = []string{"protection from Demons", "protection from Dragons"}

	plainBear := &Characteristic{Types: []string{"Creature"}, Subtypes: []string{"Bear"}}
	realDemon := &Characteristic{Types: []string{"Creature"}, Subtypes: []string{"Demon"}}
	changeling := &Characteristic{Types: []string{"Creature"}, Subtypes: []string{"Shapeshifter"}, AllCreatureTypes: true}

	if ProtectedFrom(&angel, plainBear) {
		t.Error("a Bear is neither a Demon nor a Dragon")
	}
	if !ProtectedFrom(&angel, realDemon) {
		t.Error("a Demon must be refused")
	}
	if !ProtectedFrom(&angel, changeling) {
		t.Error("a changeling is every creature type (CR 702.73a), so it is a Demon to Baneslayer Angel")
	}
	q, ok := MatchedProtection(&angel, changeling)
	if !ok || q.Printed != "Demons" {
		t.Errorf("the matched quality must be nameable for the block sentence: %+v (%v)", q, ok)
	}
}

// TestProtectionFromACardTypeReadsEffectiveTypes: "protection from
// artifacts" is about what the source IS now, not what it was printed
// as — the same layer-aware read every other check in the engine makes.
func TestProtectionFromACardTypeReadsEffectiveTypes(t *testing.T) {
	c := NewCard("Pro Artifacts", uuid.New())
	c.TypeLine = "Creature — Bear"
	c.Keywords = []string{"protection from artifacts"}

	if !ProtectedFrom(&c, &Characteristic{Types: []string{"Artifact", "Creature"}}) {
		t.Error("an artifact creature is an artifact")
	}
	if ProtectedFrom(&c, &Characteristic{Types: []string{"Creature"}}) {
		t.Error("a plain creature is not an artifact")
	}
}
