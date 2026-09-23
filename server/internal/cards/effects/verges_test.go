package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bleachboneVergeOracle = "2b8144a0-08d2-4c28-9fd7-5d90f90105e4"
	floodfarmVergeOracle  = "f1e9abfb-c3c8-483e-b446-5c2afc9f6394"
	gloomlakeVergeOracle  = "d71bda4c-3dee-4398-8fd0-f77d8743b887"
)

// The unconditional half always works; the gated half is refused —
// tapping nothing — until a land of one of the two named types is
// beside it.
func TestFloodfarmVergeSecondColourNeedsAPlainsOrAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := seedPermanentWithOracle(g, me.ID, "Floodfarm Verge", "Land", floodfarmVergeOracle)
	// A Swamp is a land, but not one the clause names.
	seedManaLand(g, me.ID, "Swamp", "Basic Land — Swamp", "B")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("gated {U} with no Plains or Island: err = %v, want ErrConditionNotMet", err)
	}
	if c, _ := battlefieldCard(g, verge); c.Tapped {
		t.Fatal("a refused activation tapped the verge")
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {W}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"W"}) {
		t.Errorf("pool = %v, want {W}", got)
	}
}

// A NONBASIC land carrying the type counts — Hallowed Fountain is a
// Plains Island — which is why the gate reads land subtypes rather than
// names.
func TestFloodfarmVergeSecondColourWithAShockland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := seedPermanentWithOracle(g, me.ID, "Floodfarm Verge", "Land", floodfarmVergeOracle)
	seedPermanentWithOracle(g, me.ID, "Hallowed Fountain", "Land — Plains Island", "f1750962-a87c-49f6-b731-02ae971ac6ea")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("gated {U} beside a Plains Island: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want {U}", got)
	}
}

// An opponent's Island is not "you control".
func TestFloodfarmVergeIgnoresAnOpponentsIsland(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	verge := seedPermanentWithOracle(g, me.ID, "Floodfarm Verge", "Land", floodfarmVergeOracle)
	seedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet", err)
	}
}

// The other two verges in the table, each switched on by one of its
// two types and producing the printed gated colour.
func TestVergeCycleGatedColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, beside, typeLine, want string
	}{
		{"Bleachbone Verge", bleachboneVergeOracle, "Plains", "Basic Land — Plains", "W"},
		{"Bleachbone Verge", bleachboneVergeOracle, "Swamp", "Basic Land — Swamp", "W"},
		{"Gloomlake Verge", gloomlakeVergeOracle, "Island", "Basic Land — Island", "B"},
		{"Gloomlake Verge", gloomlakeVergeOracle, "Swamp", "Basic Land — Swamp", "B"},
		{"Floodfarm Verge", floodfarmVergeOracle, "Island", "Basic Land — Island", "U"},
	} {
		t.Run(tc.name+"/"+tc.beside, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			verge := seedPermanentWithOracle(g, me.ID, tc.name, "Land", tc.oracle)
			seedManaLand(g, me.ID, tc.beside, tc.typeLine)
			if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("gated ability: %v", err)
			}
			if got := poolColors(me); !reflect.DeepEqual(got, []string{tc.want}) {
				t.Errorf("pool = %v, want {%s}", got, tc.want)
			}
		})
	}
}
