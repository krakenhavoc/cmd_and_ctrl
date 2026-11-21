package testsru

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gap31RishadanPort = "f3e8dc56-2810-474e-a6a7-9c3555f94ae9"
	gap31Tower        = "ac08fae8-208c-4602-8d39-9bfd29b53a5e"
	gap31Lazotep      = "ecc8e782-90a0-4a4d-b4cb-1d7383938c14"
	gap31Elevator     = "cc36ec3e-ef7e-4d3c-84f2-80ee13cbad26"
	gap31Waterways    = "271264fb-965c-4fe3-baec-2217de20f8b5"
	gap31Omenpath     = "7a82c183-1561-4276-9a75-6bbb9c2bee2e"
)

func TestGapB31CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gap31RishadanPort: "Rishadan Port",
		gap31Tower:        "Tower of the Magistrate",
		gap31Lazotep:      "Lazotep Archway",
		gap31Elevator:     "Temur Elevator",
		gap31Waterways:    "Slumbering Waterways",
		gap31Omenpath:     "Omenpath to Naya",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB31RishadanPortTapsATargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	port := effectstest.B21Push(g, me.ID, "Rishadan Port", "Land", gap31RishadanPort, 0, 0)
	theirs := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, port, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatalf("Rishadan Port targeted a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, port, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, theirs); c == nil || !c.Tapped {
		t.Errorf("the target land is not tapped")
	}
}

func TestGapB31TowerOfTheMagistrateGrantsProtectionFromArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	tower := effectstest.B21Push(g, me.ID, "Tower of the Magistrate", "Land", gap31Tower, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, bear); c == nil || !game.HasKeyword(c, "protection from artifacts") {
		t.Errorf("the Bear does not have protection from artifacts")
	}
}

func TestGapB31TemurElevatorCostsLifeWithoutTheCitysBlessing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	elevator := effectstest.B21Push(g, me.ID, "Temur Elevator", "Land", gap31Elevator, 0, 0)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, elevator, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
}

func TestGapB31TemurElevatorIsFreeWithTheCitysBlessing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	elevator := effectstest.B21Push(g, me.ID, "Temur Elevator", "Land", gap31Elevator, 0, 0)
	effectstest.GrantBlessing(g, me)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, elevator, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if me.Life != life {
		t.Errorf("life %d → %d, want unchanged", life, me.Life)
	}
}

func TestGapB31OmenpathEntersWithFourTimeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Omenpath to Naya", gap31Omenpath)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, id)
	if c == nil || c.Counters[game.CounterTime] != 4 {
		t.Fatalf("Omenpath to Naya time counters: %+v, want 4", c)
	}
}

func TestGapB31SlumberingWaterwaysEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Slumbering Waterways", gap31Waterways)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Slumbering Waterways")
}

func TestGapB31LazotepArchwayEternalizesIntoACreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	archway := effectstest.SeedGraveyardCard(t, g, "Lazotep Archway", "Land", gap31Lazotep)
	effectstest.ApaMana(me, "C", "C", "C", "W", "B")
	if err := g.ActivateCatalogAbility(me.ID, archway, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eternalize: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(archway) {
		t.Errorf("the eternalized card was not exiled")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Lazotep Archway" || !c.IsToken() {
			continue
		}
		found = true
		live := effectstest.ApaLive(g, c.InstanceID)
		if live.CurrentPower() != 4 || live.CurrentToughness() != 4 {
			t.Errorf("token %d/%d, want 4/4", live.CurrentPower(), live.CurrentToughness())
		}
		if !live.IsCreature() || live.IsLand() {
			t.Errorf("token types %q, want a creature that is not a land", live.TypeLine)
		}
		if !live.Tapped {
			t.Errorf("the token did not enter tapped")
		}
	}
	if !found {
		t.Errorf("no eternalized token")
	}
}
