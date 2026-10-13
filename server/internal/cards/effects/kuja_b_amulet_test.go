package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_b_amulet_test.go — Primal Amulet // Primal Wellspring (#3213,
// slice kuja-b).

// kbAmulet is the double-faced card as the deck importer builds it,
// front face up.
func kbAmulet(owner uuid.UUID, charge int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: primalAmuletOracleID, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Primal Amulet", TypeLine: "Artifact", ManaCost: "{4}"},
			{Name: "Primal Wellspring", TypeLine: "Land"},
		},
	}
	if charge > 0 {
		c.Counters = map[string]int{game.CounterCharge: charge}
	}
	c.SetFace(0)
	return c
}

// The discount is {1} off instants and sorceries you cast, and only those.
func TestPrimalAmuletDiscountsInstantsAndSorceries(t *testing.T) {
	seat := uuid.New()
	mods := game.CostModifiersForCard(kbAmulet(seat, 0))
	if len(mods) != 1 {
		t.Fatalf("want one cost modifier, got %d", len(mods))
	}
	m := mods[0]
	q := game.CostQuery{Controller: seat, Source: kbAmulet(seat, 0), Card: game.Card{Name: "Bolt", TypeLine: "Instant"}}
	if !m.AppliesTo(q) || m.Amount(q) != 1 {
		t.Error("an instant you cast should cost {1} less")
	}
	q.Card.TypeLine = "Creature — Bear"
	if m.AppliesTo(q) {
		t.Error("a creature spell must not be discounted")
	}
	q.Card.TypeLine = "Sorcery"
	q.Controller = uuid.New()
	if m.AppliesTo(q) {
		t.Error("an opponent's spell must not be discounted")
	}
}

// Each instant or sorcery you cast adds a charge counter; below four
// nothing else happens.
func TestPrimalAmuletCountsChargeCountersBelowFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amulet := pushBattlefieldCardWithTimestamp(g, kbAmulet(me.ID, 0))
	castCatalogSpell(t, g, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	got, _ := battlefieldCard(g, amulet)
	if n := got.Counters[game.CounterCharge]; n != 1 {
		t.Errorf("charge counters = %d, want 1", n)
	}
	if got.Name != "Primal Amulet" {
		t.Errorf("still the front face: %q", got.Name)
	}
}

// A creature spell is not an instant or sorcery.
func TestPrimalAmuletIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amulet := pushBattlefieldCardWithTimestamp(g, kbAmulet(me.ID, 0))
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	got, _ := battlefieldCard(g, amulet)
	if n := got.Counters[game.CounterCharge]; n != 0 {
		t.Errorf("charge counters = %d, want 0", n)
	}
}

// At four counters the controller may remove them and transform it, in
// place; declining keeps the counters and the front face.
func TestPrimalAmuletTransformsAtFourWhenYouAgree(t *testing.T) {
	for _, tc := range []struct {
		name string
		yes  bool
	}{{"yes", true}, {"no", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			amulet := pushBattlefieldCardWithTimestamp(g, kbAmulet(me.ID, 3))
			castCatalogSpell(t, g, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", nil)
			passPriorityAroundTable(t, g)
			answerMayChoice(t, g, me.ID, tc.yes)
			passPriorityAroundTable(t, g)
			got, ok := battlefieldCard(g, amulet)
			if !ok {
				t.Fatal("the permanent left the battlefield; it should turn over in place")
			}
			if tc.yes {
				if got.Name != "Primal Wellspring" || !got.IsLand() {
					t.Errorf("transformed into %q (%q)", got.Name, got.TypeLine)
				}
				if n := got.Counters[game.CounterCharge]; n != 0 {
					t.Errorf("charge counters = %d, want them removed", n)
				}
				return
			}
			if got.Name != "Primal Amulet" || got.Counters[game.CounterCharge] != 4 {
				t.Errorf("declined: %q with %d counters, want Primal Amulet with 4", got.Name, got.Counters[game.CounterCharge])
			}
		})
	}
}

// The Wellspring's mana, spent on an instant, copies it.
func TestPrimalWellspringCopiesAnInstantPaidWithItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	card := kbAmulet(me.ID, 0)
	card.SetFace(1)
	well := pushBattlefieldCardWithTimestamp(g, card)
	activateManaFor(t, g, me.ID, well, 0, game.ManaAbilityParams{Colors: []string{"R"}})
	before := lifeOf(g, victim.ID)

	params := strict
	params.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}
	castFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle, params)
	if triggerOnStack(g, well) == nil {
		t.Fatal("the Wellspring trigger is not on the stack after the cast")
	}
	settleKeepingCopyTargets(t, g, me.ID, victim.ID, 1)
	if got := lifeOf(g, victim.ID); got != before-6 {
		t.Errorf("victim life = %d, want %d (the Bolt and its copy)", got, before-6)
	}
}

// A creature spell paid with that mana is not copied.
func TestPrimalWellspringDoesNotCopyACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	card := kbAmulet(me.ID, 0)
	card.SetFace(1)
	well := pushBattlefieldCardWithTimestamp(g, card)
	activateManaFor(t, g, me.ID, well, 0, game.ManaAbilityParams{Colors: []string{"R"}})
	castFromHandForTest(t, g, me, "Spell", "Creature — Goblin", "{R}", "", strict)
	if triggerOnStack(g, well) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("the Wellspring triggered on a creature spell")
	}
}
