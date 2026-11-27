package game

import "testing"

// "Protection from snow" (Ronom Hulk, S82 OLD b559): a closed grammar phrase
// whose quality matches any source with the snow supertype (CR 205.4g), a
// permanent, a spell or an ability of one, and nothing else.
func TestProtectionFromSnowMatchesOnlySnowSources(t *testing.T) {
	q, ok := ParseProtectionQuality(ProtectionFromSnow)
	if !ok || q.Kind != ProtectionQualitySnow || q.Kind.String() != "snow" {
		t.Fatalf("parsed (%+v, %v), want the snow kind", q, ok)
	}
	if toks, ok := ProtectionTokens("protection from snow"); !ok || len(toks) != 1 {
		t.Errorf("ProtectionTokens = %v, %v", toks, ok)
	}
	for _, tc := range []struct {
		name  string
		src   Characteristic
		match bool
	}{
		{"a snow creature", Characteristic{Types: []string{"creature"}, Supertypes: []string{"Snow"}}, true},
		{"a snow land", Characteristic{Types: []string{"land"}, Supertypes: []string{"Basic", "Snow"}}, true},
		{"a snow enchantment", Characteristic{Types: []string{"enchantment"}, Supertypes: []string{"Snow"}}, true},
		{"a nonsnow creature", Characteristic{Types: []string{"creature"}}, false},
		{"a legendary creature", Characteristic{Types: []string{"creature"}, Supertypes: []string{"Legendary"}}, false},
	} {
		src := tc.src
		if got := q.Matches(&src); got != tc.match {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.match)
		}
	}
}
