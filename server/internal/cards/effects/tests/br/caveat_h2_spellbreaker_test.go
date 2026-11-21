package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// #3172 cav-h2: Spellbreaker Behemoth protects a creature spell whose
// power comes from its own characteristic-defining ability (CR 604.3
// works on the stack). Tarmogoyf's power is the number of card types in
// all graveyards.
func TestSpellbreakerBehemothProtectsAStarCreatureSpellOnceItsPowerIsFive(t *testing.T) {
	goyf := effectstest.CsCard("Tarmogoyf", "Creature — Lhurgoyf", "{1}{G}", 0)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.PushCatalogPermanent(g, me.ID, "Spellbreaker Behemoth", "Creature — Beast", effectstest.CsSpellbreakerOracle, false)
			g.WithWriteLock(func() {
				for _, ty := range tc.types {
					typeLine := ty
					if ty == "Land" {
						typeLine = "Basic Land — Forest"
					}
					if ty == "Creature" {
						typeLine = "Creature — Bear"
					}
					effectstest.PushGraveyardCardTyped(me, "Filler "+ty, typeLine)
				}
			})
			id := effectstest.CsPushSpell(g, effectstest.CsSpell{Card: goyf})
			if got := effectstest.CounterByEffect(t, g, id); got != tc.want {
				t.Errorf("Tarmogoyf survived the counter = %v, want %v", got, tc.want)
			}
		})
	}
}
