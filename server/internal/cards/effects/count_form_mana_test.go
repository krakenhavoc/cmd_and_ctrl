package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// count_form_mana_test.go — #3112. Mana that scales with a game quantity
// reaches the pool as a game.ManaAmount, never as a string whose length is
// that quantity: a payment-driven mana ability (ProducedForPaid) and the
// AddMana primitive both take the count form, the ceiling is applied to N
// before anything is allocated, and the cut is announced.

// A huge CountersRemoved must not allocate: the output is a count, so the
// ProducedPerCounterRemoved closure does no work proportional to N.
func TestProducedPerCounterRemovedIsACountNotAString(t *testing.T) {
	fn := ProducedPerCounterRemoved("C")
	paid := game.PaidCost{CountersRemoved: 1 << 40}
	var out game.ManaOutput
	allocs := testing.AllocsPerRun(100, func() { out = fn(nil, uuid.Nil, uuid.Nil, paid) })
	if allocs != 0 {
		t.Errorf("ProducedPerCounterRemoved at 2^40 made %d allocations, want 0", int(allocs))
	}
	if !out.IsAmount() || out.Amount.N != 1<<40 || out.Text != "" {
		t.Errorf("output = %+v, want a count of 2^40 and no text", out)
	}
	if got := fn(nil, uuid.Nil, uuid.Nil, game.PaidCost{}); got.IsAmount() && got.Amount.N != 0 {
		t.Errorf("removing nothing produced %+v, want nothing", got)
	}
}

// Mage-Ring Network cashing a pile of counters at the ceiling adds exactly
// MaxProducedMana colourless and announces the cut once.
func TestMageRingNetworkAtTheCeilingIsBoundedAndAnnounced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12Push(g, me.ID, "Mage-Ring Network", "Land", mageRingOracle, 0, 0)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, game.CounterStorage, game.MaxCount) })

	activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{game.MaxCount}})
	if got := me.ManaPool.Total(); got != game.MaxProducedMana {
		t.Errorf("pool holds %d mana, want the ceiling %d", got, game.MaxProducedMana)
	}
	if n := countCappedEvents(g, "mana"); n != 1 {
		t.Errorf("want one mana EventCountCapped, got %d", n)
	}
}

// Mana Reflection and Nyxbloom Ancient act on the count path exactly as on
// the text path: tapping Mage-Ring for three doubles / triples it.
func TestMageRingNetworkCountIsStillReplacedByManaReflectionAndNyxbloom(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		want                   int
	}{
		{"Mana Reflection", manaReflectionOracle, "Enchantment", 6},
		{"Nyxbloom Ancient", nyxbloomAncientOracle, "Enchantment Creature — Elemental", 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			land := b12Push(g, me.ID, "Mage-Ring Network", "Land", mageRingOracle, 0, 0)
			advanceToMain(t, g)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, game.CounterStorage, 4) })

			activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{3}})
			if got := me.ManaPool.Total(); got != tc.want {
				t.Errorf("pool holds %d mana, want %d", got, tc.want)
			}
			if len(g.PendingChoices) != 0 {
				t.Errorf("queued %d prompts, want none", len(g.PendingChoices))
			}
		})
	}
}

func addManaAmount(t *testing.T, g *game.Game, a AddMana) {
	t.Helper()
	me := g.Seats[0]
	src := uuid.New()
	if err := a.Apply(NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: src})); err != nil {
		t.Fatalf("AddMana.Apply: %v", err)
	}
}

func TestAddManaCountFormAddsExactlyN(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	addManaAmount(t, g, AddMana{Amount: ManaN(3, "R")})
	if got := sortedPool(me); len(got) != 3 || got[0] != "R" || got[2] != "R" {
		t.Errorf("pool = %v, want three R", got)
	}
	// Zero and negative counts add nothing and do not error.
	addManaAmount(t, g, AddMana{Amount: ManaN(0, "R")})
	addManaAmount(t, g, AddMana{Amount: ManaN(-4, "R")})
	if me.ManaPool.Total() != 3 {
		t.Errorf("pool grew to %d on a non-positive count", me.ManaPool.Total())
	}
}

// The ceiling on the AddMana path: 2^40 adds MaxProducedMana and says so.
func TestAddManaCountFormIsCappedAndAnnounced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	addManaAmount(t, g, AddMana{Amount: ManaN(1<<40, "R")})
	if got := me.ManaPool.Total(); got != game.MaxProducedMana {
		t.Errorf("pool holds %d, want the ceiling %d", got, game.MaxProducedMana)
	}
	if n := countCappedEvents(g, "mana"); n != 1 {
		t.Errorf("want one mana EventCountCapped, got %d", n)
	}
	// Exactly at the ceiling is not a cut.
	g2 := newCatalogGame(t)
	addManaAmount(t, g2, AddMana{Amount: ManaN(game.MaxProducedMana, "R")})
	if n := countCappedEvents(g2, "mana"); n != 0 {
		t.Errorf("a count AT the ceiling announced a cut (%d events)", n)
	}
}

// Spend restrictions ride the count form (Karn, Metamorphosis).
func TestAddManaCountFormCarriesRestrictions(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	addManaAmount(t, g, AddMana{Amount: ManaN(2, "G"), Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")}})
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %d, want 2", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool.Units() {
		if len(tok.Restrictions) != 2 {
			t.Errorf("token restrictions = %v, want the two declared", tok.Restrictions)
		}
	}
}

// A multi-colour option is N independent picks (Burnt Offering, Culling
// Ritual), and a pick cannot carry a restriction.
func TestAddManaCountFormQueuesOnePickPerMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	addManaAmount(t, g, AddMana{Amount: ManaN(3, "B", "R")})
	picks := 0
	for _, pc := range g.PendingChoices {
		if pc.Kind == game.PendingChoiceMana && pc.Chooser == me.ID {
			picks++
		}
	}
	if picks != 3 {
		t.Errorf("queued %d colour picks, want 3", picks)
	}
	err := AddMana{Amount: ManaN(1, "B", "R"), Restrictions: []string{ManaRestrictCast}}.Apply(
		NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}))
	if err == nil {
		t.Error("a restricted colour pick was accepted; it would mint unrestricted mana")
	}
}

// A spell's mana is not a tap, so Mana Reflection leaves the count form
// alone, as it leaves the text form alone.
func TestAddManaCountFormIsNotDoubledByManaReflection(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mana Reflection", "Enchantment", manaReflectionOracle, false)
	addManaAmount(t, g, AddMana{Amount: ManaN(3, "B")})
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("a spell's counted mana made %d, want 3 (nothing was tapped)", got)
	}
}

// Both forms agree: the count form and the text form mint the same pool.
func TestAddManaCountFormMatchesTheTextForm(t *testing.T) {
	a, b := newCatalogGame(t), newCatalogGame(t)
	addManaAmount(t, a, AddMana{Amount: ManaN(4, "C")})
	addManaAmount(t, b, AddMana{Produced: "{C}{C}{C}{C}"})
	pa, pb := sortedPool(a.Seats[0]), sortedPool(b.Seats[0])
	if len(pa) != len(pb) {
		t.Fatalf("count form %v vs text form %v", pa, pb)
	}
	for i := range pa {
		if pa[i] != pb[i] {
			t.Fatalf("count form %v vs text form %v", pa, pb)
		}
	}
}
