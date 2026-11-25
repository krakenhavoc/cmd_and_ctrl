package game

import "testing"

// "Protection from legendary creatures" (Tsabo Tavoc, #5462): a closed
// grammar phrase whose quality matches a creature source with the legendary
// supertype, and nothing else.
func TestProtectionFromLegendaryCreaturesMatchesOnlyLegendaryCreatureSources(t *testing.T) {
	q, ok := ParseProtectionQuality(ProtectionFromLegendaryCreatures)
	if !ok || q.Kind != ProtectionQualityLegendaryCreatures || q.Kind.String() != "legendary_creatures" {
		t.Fatalf("parsed (%+v, %v), want the legendary_creatures kind", q, ok)
	}
	if toks, ok := ProtectionTokens("protection from legendary creatures"); !ok || len(toks) != 1 {
		t.Errorf("ProtectionTokens = %v, %v", toks, ok)
	}
	for _, tc := range []struct {
		name  string
		src   Characteristic
		match bool
	}{
		{"a legendary creature", Characteristic{Types: []string{"creature"}, Supertypes: []string{"Legendary"}}, true},
		{"a legendary artifact creature", Characteristic{Types: []string{"artifact", "creature"}, Supertypes: []string{"Legendary"}}, true},
		{"a nonlegendary creature", Characteristic{Types: []string{"creature"}}, false},
		{"a legendary artifact", Characteristic{Types: []string{"artifact"}, Supertypes: []string{"Legendary"}}, false},
		{"a legendary planeswalker", Characteristic{Types: []string{"planeswalker"}, Supertypes: []string{"Legendary"}}, false},
	} {
		src := tc.src
		if got := q.Matches(&src); got != tc.match {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.match)
		}
	}
	for _, bad := range []string{"protection from legendary", "protection from legendary permanents", "protection from legendary creature"} {
		if _, ok := ParseProtectionQuality(bad); ok {
			t.Errorf("%q parsed; the grammar is closed", bad)
		}
	}
}
