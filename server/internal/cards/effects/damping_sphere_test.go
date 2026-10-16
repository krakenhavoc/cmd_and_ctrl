package effects

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damping_sphere_test.go — #3939: "If a land is tapped for two or more
// mana, it produces {C} instead of any other type and amount." The
// production window sees a mana ability's whole output (CR 106.12b), so
// the Sphere can count it.

func dampingSphereOn(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Damping Sphere", "Artifact", dampingSphereOracle, false)
}

func sortedColors(p *game.Player) []string {
	out := batch01PoolColors(p)
	sort.Strings(out)
	return out
}

// Ancient Tomb's {C}{C} is two mana from one land: one {C}. The rider
// still deals its 2 damage, which the Sphere does not touch.
func TestDampingSphereMakesAncientTombProduceOneColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dampingSphereOn(g, me.ID)
	tomb := seedPermanentWithOracle(g, me.ID, "Ancient Tomb", "Land", ancientTombOracle)
	life := me.Life

	tapForMana(t, g, me.ID, tomb)
	if got := sortedColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want [C]", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d (the land still deals 2)", me.Life, life-2)
	}
}

// A bounce land's {W}{U} is two mana of two colours: {C} "instead of
// any other type and amount".
func TestDampingSphereMakesABounceLandProduceOneColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dampingSphereOn(g, me.ID)
	chancery := seedPermanentWithOracle(g, me.ID, "Azorius Chancery", "Land", azoriusChanceryOracle)

	tapForMana(t, g, me.ID, chancery)
	if got := sortedColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want [C]", got)
	}
}

// A basic land taps for one mana and is left alone.
func TestDampingSphereLeavesABasicLandAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dampingSphereOn(g, me.ID)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")

	tapForMana(t, g, me.ID, forest)
	if got := sortedColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want [G]", got)
	}
}

// The Sphere is symmetric: an opponent's land is caught too.
func TestDampingSphereCatchesAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dampingSphereOn(g, me.ID)
	tomb := seedPermanentWithOracle(g, opp.ID, "Ancient Tomb", "Land", ancientTombOracle)

	tapForMana(t, g, opp.ID, tomb)
	if got := sortedColors(opp); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("opponent's pool = %v, want [C]", got)
	}
}

// A filter land's two picks are one production: both are asked before
// any mana arrives, and the Sphere then makes the pair one {C}. Without
// the Sphere the same two answers make the two colours picked.
func TestDampingSphereCatchesAFilterLandsTwoPicks(t *testing.T) {
	for _, sphere := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		if sphere {
			dampingSphereOn(g, me.ID)
		}
		gate := seedPermanentWithOracle(g, me.ID, "Mystic Gate", "Land", b07MysticGateOracle)
		me.ManaPool.AddMana(game.ManaToken{Color: "U"})
		if err := g.ActivateManaAbility(me.ID, gate, 1, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		for i, color := range []string{"W", "U"} {
			pick := riderLatestManaPick(g, me.ID)
			if pick == nil {
				t.Fatalf("sphere=%v: no pick open for answer %d", sphere, i+1)
			}
			if me.ManaPool.Total() != 0 {
				t.Errorf("sphere=%v: mana arrived before the last pick: %v", sphere, batch01PoolColors(me))
			}
			if err := g.ResolveManaChoice(pick.ID, me.ID, color); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
		if pick := riderLatestManaPick(g, me.ID); pick != nil {
			t.Errorf("sphere=%v: a third pick is open: %+v", sphere, pick)
		}
		want := []string{"U", "W"}
		if sphere {
			want = []string{"C"}
		}
		if got := sortedColors(me); !reflect.DeepEqual(got, want) {
			t.Errorf("sphere=%v: pool = %v, want %v", sphere, got, want)
		}
	}
}

// "Two or more" is read as the event stands (CR 616.1): Mana Reflection
// makes a Forest's {G} into {G}{G}, and the Sphere then applies to two
// mana.
func TestDampingSphereAppliesAfterManaReflectionDoublesABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dampingSphereOn(g, me.ID)
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")

	tapForMana(t, g, me.ID, forest)
	if got := sortedColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool = %v, want [C]", got)
	}
}

// The auto-tapper prices the whole output: under the Sphere, Ancient
// Tomb is one mana, so it alone cannot pay {2}, and with a Forest beside
// it the cast taps both and pays.
func TestDampingSphereIsPricedByTheAutoTapper(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	tomb := seedPermanentWithOracle(g, me.ID, "Ancient Tomb", "Land", ancientTombOracle)
	cost, err := game.ParseCost("{2}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); !ok || len(plan) != 1 || plan[0] != tomb {
		t.Fatalf("without the Sphere, Ancient Tomb alone pays {2}: plan %v, ok %v", plan, ok)
	}
	dampingSphereOn(g, me.ID)
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatalf("under the Sphere Ancient Tomb makes one mana, but the planner paid {2} with %v", plan)
	}

	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 2 {
		t.Fatalf("Tomb and Forest pay {2}: plan %v, ok %v", plan, ok)
	}
	id := handCardFull(me, "Mind Stone", "Artifact", "{2}", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"}); err != nil {
		t.Fatalf("auto-tapped cast of {2}: %v", err)
	}
	for _, src := range []uuid.UUID{tomb, forest} {
		if c, _ := battlefieldCard(g, src); !c.Tapped {
			t.Errorf("%v should have been tapped to pay", src)
		}
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("%d mana floating after the cast, want 0: %v", left, batch01PoolColors(me))
	}
}
