package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// set_creature_types_cards_test.go — the cards on the set-creature-types
// seam (#2144): Polymorphist's Jest (the Sauron deck request, #2062)
// and Frodo, Sauron's Bane.

const (
	polymorphistsJestOracle = "5bc66b2d-a93c-449d-a090-b61e1613aa15"
	frodoSauronsBaneOracle  = "8be62f5c-0ab2-4869-97a2-d8a69a3fde95"
	frodoSauronsBaneType    = "Legendary Creature — Halfling Citizen"
)

// effectiveCardForTest is the card with its layers applied.
func effectiveCardForTest(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	ok := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok = g.LookupCardForEffect(id)
	})
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func TestSetCreatureTypesCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		polymorphistsJestOracle: "Polymorphist's Jest",
		frodoSauronsBaneOracle:  "Frodo, Sauron's Bane",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s registered as %q, %v; want Full", oracle, spec.Name, spec.Completeness)
		}
	}
}

// --- Polymorphist's Jest --------------------------------------------

// Each creature the target player controls is a blue 1/1 Frog with no
// abilities; an artifact creature is still an artifact; the caster's
// own creature is untouched.
func TestPolymorphistsJestMakesTheirCreaturesBlueFrogs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shivan Dragon", TypeLine: "Creature — Dragon", Colors: []string{"R"},
		Power: 5, Toughness: 5, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Iron Golem", TypeLine: "Artifact Creature — Golem",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Polymorphist's Jest", "Instant", polymorphistsJestOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	d := effectiveCardForTest(t, g, dragon)
	if ch := d.Effective(); !reflect.DeepEqual(ch.Subtypes, []string{"Frog"}) || !reflect.DeepEqual(ch.Colors, []string{"U"}) {
		t.Errorf("dragon is %v %v, want a blue Frog", ch.Colors, ch.Subtypes)
	}
	if effectiveHasKeyword(&d, "flying") {
		t.Error("the dragon kept flying")
	}
	if p, tough := effectivePower(t, g, dragon), effectiveToughness(t, g, dragon); p != 1 || tough != 1 {
		t.Errorf("dragon is %d/%d, want 1/1", p, tough)
	}
	gc := effectiveCardForTest(t, g, golem).Effective()
	if !reflect.DeepEqual(gc.Types, []string{"Artifact", "Creature"}) || !reflect.DeepEqual(gc.Subtypes, []string{"Frog"}) {
		t.Errorf("golem is %v — %v, want Artifact Creature — Frog", gc.Types, gc.Subtypes)
	}
	if mc := effectiveCardForTest(t, g, mine).Effective(); !reflect.DeepEqual(mc.Subtypes, []string{"Bear"}) {
		t.Errorf("the caster's Bear became %v", mc.Subtypes)
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("the caster's Bear has power %d, want 2", p)
	}
}

// --- Frodo, Sauron's Bane -------------------------------------------

func pushFrodoSauronsBane(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := pushCatalogPermanent(g, owner, "Frodo, Sauron's Bane", frodoSauronsBaneType, frodoSauronsBaneOracle, false)
	if c := findBattlefieldCardByID(g, id); c != nil {
		c.Power, c.Toughness, c.PrintedPTKnown = 1, 2, true
	}
	return id
}

func activateFrodo(t *testing.T, g *game.Game, me *game.Player, frodo uuid.UUID, idx int) {
	t.Helper()
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, frodo, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
	passPriorityAroundTable(t, g)
}

// The Scout ability works on a Citizen; the Rogue ability does nothing
// until Frodo is a Scout, and then makes him a Rogue who keeps the
// Scout's 2/3 and lifelink and is no longer a Scout or a Citizen.
func TestFrodoSauronsBaneClimbsFromCitizenToScoutToRogue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	frodo := pushFrodoSauronsBane(g, me.ID)

	activateFrodo(t, g, me, frodo, 1)
	if c := effectiveCardForTest(t, g, frodo); c.HasSubtype("Rogue") || !c.HasSubtype("Citizen") {
		t.Fatal("the Rogue ability changed a Citizen")
	}

	activateFrodo(t, g, me, frodo, 0)
	scout := effectiveCardForTest(t, g, frodo)
	if !scout.HasSubtype("Scout") || !scout.HasSubtype("Halfling") || scout.HasSubtype("Citizen") {
		t.Errorf("after the first ability Frodo is %v, want Halfling Scout", scout.Effective().Subtypes)
	}
	if !effectiveHasKeyword(&scout, "lifelink") {
		t.Error("the Scout has no lifelink")
	}
	if p, tough := effectivePower(t, g, frodo), effectiveToughness(t, g, frodo); p != 2 || tough != 3 {
		t.Errorf("the Scout is %d/%d, want 2/3", p, tough)
	}

	activateFrodo(t, g, me, frodo, 1)
	rogue := effectiveCardForTest(t, g, frodo)
	if !reflect.DeepEqual(rogue.Effective().Subtypes, []string{"Halfling", "Rogue"}) {
		t.Errorf("after the second ability Frodo is %v, want Halfling Rogue", rogue.Effective().Subtypes)
	}
	if !effectiveHasKeyword(&rogue, "lifelink") {
		t.Error("the Rogue lost the Scout's lifelink")
	}
	if p, tough := effectivePower(t, g, frodo), effectiveToughness(t, g, frodo); p != 2 || tough != 3 {
		t.Errorf("the Rogue is %d/%d, want the Scout's 2/3", p, tough)
	}

	// A Rogue is not a Citizen, so the first ability does nothing now.
	activateFrodo(t, g, me, frodo, 0)
	if c := effectiveCardForTest(t, g, frodo); !c.HasSubtype("Rogue") || c.HasSubtype("Scout") {
		t.Errorf("the Scout ability changed a Rogue into %v", c.Effective().Subtypes)
	}
}

// rogueFrodoHits makes Frodo a Rogue, has the Ring tempt his controller
// `tempts` times first (through another creature, so Frodo is not the
// Ring-bearer), and has him deal combat damage to the next seat.
func rogueFrodoHits(t *testing.T, tempts int) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if tempts > 0 {
		b12Creature(g, me.ID, "Ring Bear", "Creature — Bear", 2, 2)
		ringTemptTimes(t, g, me.ID, tempts)
	}
	toMain(t, g)
	frodo := pushFrodoSauronsBane(g, me.ID)
	activateFrodo(t, g, me, frodo, 0)
	activateFrodo(t, g, me, frodo, 1)
	attackWith(t, g, opp.ID, frodo)
	ringSettle(t, g)
	return g, me, opp
}

// Under four temptations, the hit has the Ring tempt Frodo's
// controller, and nobody loses.
func TestFrodoSauronsBaneRogueTemptsYouBeforeTheFourth(t *testing.T) {
	g, me, opp := rogueFrodoHits(t, 0)
	var n int
	g.ReadSnapshot(func() { n = game.RingTemptCount(g, me.ID) })
	if n != 1 {
		t.Errorf("the Ring has tempted Frodo's controller %d times, want 1", n)
	}
	if opp.Eliminated {
		t.Error("the damaged player lost with the Ring at no temptations")
	}
}

// From the fourth temptation on, the damaged player loses the game and
// the Ring does not tempt anyone.
func TestFrodoSauronsBaneRogueEndsThePlayerHeHitsAfterFourTemptations(t *testing.T) {
	g, me, opp := rogueFrodoHits(t, 4)
	if !opp.Eliminated {
		t.Fatal("the damaged player is still in the game")
	}
	var n int
	g.ReadSnapshot(func() { n = game.RingTemptCount(g, me.ID) })
	if n != 4 {
		t.Errorf("the Ring tempted again: %d temptations, want 4", n)
	}
	for _, p := range g.Seats {
		if p.ID != opp.ID && p.Eliminated {
			t.Errorf("%s lost too", p.Name)
		}
	}
}
