package game

import "testing"

// #4505: CastPermission.ExileInsteadOnlyInstantOrSorcery narrows the
// "exile it instead" a granted cast carries to instants and sorceries
// (Bilbo, Thief in the Night). Without the narrowing every spell cast
// through the permission carries it, as Torrential Gearhulk's and
// Emet-Selch's do.
func TestExileInsteadOnlyInstantOrSorceryNarrowsTheReplacement(t *testing.T) {
	cards := map[string]Card{
		"instant":  {Name: "Opt", TypeLine: "Instant", ManaCost: "{U}"},
		"sorcery":  {Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
		"artifact": {Name: "Relic", TypeLine: "Artifact", ManaCost: "{2}"},
		"artifact instant": {Name: "Artifact Instant", TypeLine: "Artifact Instant",
			ManaCost: "{1}"},
	}
	for _, tc := range []struct {
		card   string
		narrow bool
		want   bool
	}{
		{"instant", true, true},
		{"sorcery", true, true},
		{"artifact", true, false},
		{"artifact instant", true, true},
		{"artifact", false, true},
		{"instant", false, true},
	} {
		p := &CastPermission{
			Zone:                             ZoneGraveyard,
			AltCostKey:                       AltCostKeyGrantedCost,
			ExileInsteadOfGraveyard:          true,
			ExileInsteadOnlyInstantOrSorcery: tc.narrow,
		}
		alt := p.AlternativeCostFor(cards[tc.card])
		if alt == nil {
			t.Fatalf("%s: no alternative cost", tc.card)
		}
		if alt.ExileInsteadOfGraveyard != tc.want {
			t.Errorf("%s (narrowed %v): ExileInsteadOfGraveyard = %v, want %v",
				tc.card, tc.narrow, alt.ExileInsteadOfGraveyard, tc.want)
		}
	}
	// The narrowing alone exiles nothing.
	p := &CastPermission{Zone: ZoneGraveyard, AltCostKey: AltCostKeyGrantedCost, ExileInsteadOnlyInstantOrSorcery: true}
	if alt := p.AlternativeCostFor(cards["instant"]); alt == nil || alt.ExileInsteadOfGraveyard {
		t.Error("the narrowing without ExileInsteadOfGraveyard exiled an instant")
	}
}
