package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_granted_test.go — ADR 0093 Decision 6 and its owner decision
// (2026-09-24). The auto-tapper offers every acceptable mana ability of
// a permanent as mutually exclusive candidates (so a granted
// any-colour ability can pay where the permanent's own could not), and
// it plans a CREATURE's granted mana only as a last resort — in the
// tier #1215 gave Treasures — while a land's granted mana is an
// ordinary source.

const (
	atRiteOracle    = "autotap-fixture-rite"    // creatures you control have "{T}: Add one mana of any color."
	atLanternOracle = "autotap-fixture-lantern" // lands you control have "{T}: Add one mana of any color."
	atAnyColor      = "autotap-fixture/any-color"
	atElfOracle     = "autotap-fixture-elf" // "{T}: Add {G}."
)

func stubAutoTapGrants(t *testing.T) {
	t.Helper()
	grant := func(pred func(target *Card, source *Card) bool) *CardDef {
		return &CardDef{Static: []StaticAbility{{
			Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.Controller == source.Controller && pred(target, source)
			},
			GrantAbilities: []string{atAnyColor},
		}}}
	}
	defs := map[string]*CardDef{
		atRiteOracle:    grant(func(target, _ *Card) bool { return target.IsCreature() }),
		atLanternOracle: grant(func(target, _ *Card) bool { return target.IsLand() }),
		atElfOracle:     {ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}},
		GrantKey(atAnyColor): {
			ManaAbilities: []ManaAbilityShape{{TapCost: true, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
			GrantText:     "{T}: Add one mana of any color.",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

// atPush puts a permanent on the battlefield that has been there since
// the turn began, so a {T} is not sick.
func atPush(t *testing.T, g *Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := pushTypedTestCard(g, Card{Name: name, TypeLine: typeLine, OracleID: oracle, Owner: owner, Controller: owner})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func planIDs(t *testing.T, g *Game, owner uuid.UUID, cost string) ([]uuid.UUID, bool) {
	t.Helper()
	return g.AutoTapForCost(owner, costFor(t, cost), 0)
}

// An elf under the Rite pays {U}: the granted any-colour ability is a
// candidate beside the elf's own {G}, not hidden behind it.
func TestAutoTapUsesAGrantedAbilityTheOwnOneCannotPay(t *testing.T) {
	stubAutoTapGrants(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elf := atPush(t, g, me, "Elf", "Creature — Elf Druid", atElfOracle)
	if _, ok := planIDs(t, g, me, "{U}"); ok {
		t.Fatal("fixture: without the Rite the elf cannot pay {U}")
	}
	atPush(t, g, me, "Rite", "Enchantment", atRiteOracle)
	plan, ok := planIDs(t, g, me, "{U}")
	if !ok || len(plan) != 1 || plan[0] != elf {
		t.Fatalf("plan = %v (%v), want the elf for its granted any-colour ability", plan, ok)
	}
	// And the executor fires the ability the plan booked: the pool
	// gets a blue mana, not a green one.
	g.WithWriteLock(func() {
		p := g.playerByIDLocked(me)
		cost := costFor(t, "{U}")
		booked, ok := g.autoTapLocked(me, cost, 0, nil)
		if !ok {
			t.Fatal("no plan under the lock")
		}
		g.materializePlanLocked(p, booked, cost)
		if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "U" {
			t.Errorf("pool = %v, want one {U}", p.ManaPool)
		}
	})
}

// A cost the lands can pay taps no creature for granted mana.
func TestAutoTapPrefersLandsOverCreaturesGrantedMana(t *testing.T) {
	stubAutoTapGrants(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	f1 := atPush(t, g, me, "Forest", "Basic Land — Forest", "")
	f2 := atPush(t, g, me, "Forest", "Basic Land — Forest", "")
	bear1 := atPush(t, g, me, "Bear", "Creature — Bear", "")
	bear2 := atPush(t, g, me, "Bear", "Creature — Bear", "")
	atPush(t, g, me, "Rite", "Enchantment", atRiteOracle)
	plan, ok := planIDs(t, g, me, "{1}{G}")
	if !ok {
		t.Fatal("{1}{G} should be payable")
	}
	if !containsID(plan, f1) || !containsID(plan, f2) || containsID(plan, bear1) || containsID(plan, bear2) {
		t.Errorf("plan = %v, want the two Forests and no creature", plan)
	}
}

// A creature's granted mana is used only when nothing else can pay.
func TestAutoTapUsesCreaturesGrantedManaAsALastResort(t *testing.T) {
	stubAutoTapGrants(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	forest := atPush(t, g, me, "Forest", "Basic Land — Forest", "")
	bear := atPush(t, g, me, "Bear", "Creature — Bear", "")
	atPush(t, g, me, "Rite", "Enchantment", atRiteOracle)
	plan, ok := planIDs(t, g, me, "{G}{U}")
	if !ok || !containsID(plan, forest) || !containsID(plan, bear) {
		t.Fatalf("plan = %v (%v), want the Forest for {G} and the bear's granted mana for {U}", plan, ok)
	}
	// And a creature that entered this turn cannot pay (CR 302.6).
	setSummonedThisTurn(g, bear, true)
	g.BumpLayerVersionForTest()
	if plan, ok := planIDs(t, g, me, "{G}{U}"); ok {
		t.Errorf("planned %v off a summoning-sick creature's granted {T}", plan)
	}
}

// A creature's OWN mana ability is an ordinary source, as it always
// was: an Elf pays {G} ahead of nothing being held back for it.
func TestAutoTapCreaturesOwnManaIsStillOrdinary(t *testing.T) {
	stubAutoTapGrants(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	elf := atPush(t, g, me, "Elf", "Creature — Elf Druid", atElfOracle)
	bear := atPush(t, g, me, "Bear", "Creature — Bear", "")
	atPush(t, g, me, "Rite", "Enchantment", atRiteOracle)
	plan, ok := planIDs(t, g, me, "{G}")
	if !ok || len(plan) != 1 || plan[0] != elf {
		t.Errorf("plan = %v (%v), want the elf's own {G}, not the bear's granted mana", plan, ok)
	}
	_ = bear
}

// A land's granted mana is an ordinary source: a Forest under the
// Lantern pays {U} ahead of a creature under the Rite.
func TestAutoTapPlansALandsGrantedManaNormally(t *testing.T) {
	stubAutoTapGrants(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	forest := atPush(t, g, me, "Forest", "Basic Land — Forest", "")
	bear := atPush(t, g, me, "Bear", "Creature — Bear", "")
	atPush(t, g, me, "Rite", "Enchantment", atRiteOracle)
	atPush(t, g, me, "Lantern", "Artifact", atLanternOracle)
	plan, ok := planIDs(t, g, me, "{U}")
	if !ok || len(plan) != 1 || plan[0] != forest {
		t.Errorf("plan = %v (%v), want the Lantern'd Forest, not the bear", plan, ok)
	}
	_ = bear
}

// Every acceptable ability is a candidate, not only the first — which
// also fixes an uncatalogued dual land (two intrinsic land types) that
// the planner used to read as its first colour only.
func TestAutoTapOffersEveryIntrinsicLandAbility(t *testing.T) {
	stubAutoTapGrants(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bayou := atPush(t, g, me, "Bayou", "Land — Swamp Forest", "")
	plan, ok := planIDs(t, g, me, "{G}")
	if !ok || len(plan) != 1 || plan[0] != bayou {
		t.Errorf("plan = %v (%v), want the Bayou for its {G}", plan, ok)
	}
	// One permanent is tapped once: two pips need two sources.
	if plan, ok := planIDs(t, g, me, "{B}{G}"); ok {
		t.Errorf("planned %v — one Bayou cannot pay two pips", plan)
	}
}
