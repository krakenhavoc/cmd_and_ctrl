package effects

import "testing"

const spectrumSentinelOracle = "7c13d0ad-1235-41f7-8f9a-f39d7c544b74"

func TestSpectrumSentinelIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(spectrumSentinelOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 ||
		!hasKeyword(spec.PrintedKeywords, "protection from multicolored") {
		t.Fatalf("registered %v, spec %+v", ok, spec)
	}
}

// A nonbasic land an opponent controls entering gains 1 life; a basic
// land, and the controller's own nonbasic land, gain nothing.
func TestSpectrumSentinelGainsLifeForOpponentNonbasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Spectrum Sentinel", "Artifact Creature — Soldier", spectrumSentinelOracle, false)
	life := me.Life

	b34PlayLandAs(t, g, 1, "Swamp", "Basic Land — Swamp")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("an opponent's basic land gained life: %d → %d", life, me.Life)
	}

	b34LandEntersByEffect(t, g, me, "My Nonbasic", "Land")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("your own nonbasic land gained life: %d → %d", life, me.Life)
	}

	b34LandEntersByEffect(t, g, opp, "Their Nonbasic", "Land")
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("an opponent's nonbasic land: life %d → %d, want +1", life, me.Life)
	}
	b34PlayLandAs(t, g, 2, "Cave", "Land")
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("a second opponent's nonbasic land: life %d → %d, want +2", life, me.Life)
	}
}
