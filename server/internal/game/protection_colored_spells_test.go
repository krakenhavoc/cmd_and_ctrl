package game

import "testing"

// protection_colored_spells_test.go — "protection from spells that are one
// or more colors" (Emrakul, the Aeons Torn, CR 702.16a): only a SPELL with at
// least one colour is stopped.

func TestProtectionFromColoredSpellsGrammar(t *testing.T) {
	const token = "protection from spells that are one or more colors"
	q, ok := ParseProtectionQuality(token)
	if !ok || q.Kind != ProtectionQualityColoredSpells || q.Kind.String() != "colored_spells" {
		t.Fatalf("%q = (%+v, %v)", token, q, ok)
	}
	if _, ok := CanonicalKeywords(token); !ok {
		t.Errorf("CanonicalKeywords refused %q", token)
	}
}

func TestProtectionFromColoredSpellsMatchesOnlyAColoredSpell(t *testing.T) {
	q := ProtectionQuality{Kind: ProtectionQualityColoredSpells}
	for _, tc := range []struct {
		name string
		src  *Characteristic
		want bool
	}{
		{"red spell", &Characteristic{Colors: []string{"R"}, SourceIsSpell: true}, true},
		{"gold spell", &Characteristic{Colors: []string{"B", "G"}, SourceIsSpell: true}, true},
		{"colorless spell", &Characteristic{SourceIsSpell: true}, false},
		{"red ability or permanent", &Characteristic{Colors: []string{"R"}, SourceIsPermanent: true}, false},
		{"unknown source", nil, false},
	} {
		if got := q.Matches(tc.src); got != tc.want {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCountersRemainOnACardThatSaysSoExceptInAHandOrLibrary(t *testing.T) {
	RegisterCountersRemain("test-counters-remain")
	if !countersRemainOn("test-counters-remain") || countersRemainOn("") || countersRemainOn("other") {
		t.Fatal("countersRemainOn does not read the registry")
	}
}
