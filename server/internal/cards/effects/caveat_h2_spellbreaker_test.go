package effects

import (
	"testing"
)

// #3172 cav-h2: Spellbreaker Behemoth protects a creature spell whose
// power comes from its own characteristic-defining ability (CR 604.3
// works on the stack). Tarmogoyf's power is the number of card types in
// all graveyards.
func TestSpellbreakerBehemothProtectsAStarCreatureSpellOnceItsPowerIsFive(t *testing.T) {
	goyf := csCard("Tarmogoyf", "Creature — Lhurgoyf", "{1}{G}", 0)
	goyf.OracleID = "45900b2f-f6a9-4c42-9642-008f3c1cf6dd"
	goyf.Toughness = 0
	for _, tc := range []struct {
		name  string
		types []string
		want  bool
	}{
		{"four types is power 4", []string{"Artifact", "Creature", "Land", "Instant"}, false},
		{"five types is power 5", []string{"Artifact", "Creature", "Land", "Instant", "Sorcery"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pushCatalogPermanent(g, me.ID, "Spellbreaker Behemoth", "Creature — Beast", csSpellbreakerOracle, false)
			g.WithWriteLock(func() {
				for _, ty := range tc.types {
					typeLine := ty
					if ty == "Land" {
						typeLine = "Basic Land — Forest"
					}
					if ty == "Creature" {
						typeLine = "Creature — Bear"
					}
					pushGraveyardCardTyped(me, "Filler "+ty, typeLine)
				}
			})
			id := csPushSpell(g, csSpell{card: goyf})
			if got := counterByEffect(t, g, id); got != tc.want {
				t.Errorf("Tarmogoyf survived the counter = %v, want %v", got, tc.want)
			}
		})
	}
}
