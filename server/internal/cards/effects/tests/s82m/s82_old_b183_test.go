package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b183_test.go — Ravnica Allegiance leftovers (S82 older sets, batch 183).

const (
	s82Old183ShieldOracle = "9c0aba12-4460-4150-b74f-93f5f8006cf5"
	s82Old183NikyaOracle  = "ebf3fd80-9574-489b-9428-f4251834bdc4"
)

func TestS82Old183CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82Old183ShieldOracle: "Screaming Shield",
		s82Old183NikyaOracle:  "Nikya of the Old Ways",
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

func TestS82Old183ScreamingShieldGivesToughnessAndMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	shield := effectstest.B21Push(g, me.ID, "Screaming Shield", "Artifact — Equipment", s82Old183ShieldOracle, 0, 0)
	effectstest.SeedLibrary(opp, "One", "Two", "Three", "Four")
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, shield, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("equipped power %d, want 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 5 {
		t.Errorf("equipped toughness %d, want 5", got)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has the granted ability")
	}
	params := game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, params); err == nil {
		t.Fatal("activated with no mana")
	}
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, bear, idx, params)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != library-3 {
		t.Errorf("library %d → %d, want -3", library, got)
	}
	if got := opp.Graveyard.Size(); got != graveyard+3 {
		t.Errorf("graveyard %d → %d, want +3", graveyard, got)
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature taps to pay")
	}
}

func TestS82Old183NikyaDoublesALandsManaAndBarsNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B21Push(g, me.ID, "Nikya of the Old Ways", "Legendary Creature — Centaur Druid", s82Old183NikyaOracle, 5, 5)
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Divination", "Sorcery", "", nil); err == nil {
		t.Error("cast a noncreature spell with Nikya out")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Grizzly Bears", "Creature — Bear", "", nil); err != nil {
		t.Errorf("a creature spell is refused: %v", err)
	}
}
