package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// improvise_test.go — improvise (CR 702.126) and the grant of it: Kappa
// Cannoneer's printed improvise and Archway of Innovation's "the next
// spell you cast this turn has improvise" (#3654).

const (
	kappaCannoneerOracle = "7fffbe7f-ad78-4bb6-a611-c37e495c6ee0"
	archwayOracle        = "bfa20bc7-4626-4a52-87f4-6e2763cb8ed5"
)

func impArtifacts(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = pushTapCostPermanent(g, owner, "Trinket", "Artifact")
	}
	return ids
}

func impBear(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: owner, Controller: owner}
}

func impKappa(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Kappa Cannoneer", TypeLine: "Artifact Creature — Turtle Warrior",
		OracleID: kappaCannoneerOracle, ManaCost: "{5}{U}", Colors: []string{"U"}, Power: 4, Toughness: 4,
		Owner: owner, Controller: owner}
}

func TestKappaCannoneerImprovisePaysGenericButNotBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 5)
	kappa := impKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })

	// Five artifacts pay {5}; the {U} still needs mana.
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Fatal("cast with no blue mana succeeded: an artifact must not pay {U}")
	}
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatalf("improvising Kappa Cannoneer: %v", err)
	}
	if !cgTapped(g, arts) {
		t.Error("the improvising artifacts were not tapped")
	}
}

func TestKappaCannoneerRefusesMoreArtifactsThanGeneric(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 6)
	kappa := impKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Error("six artifacts were accepted for five generic")
	}
}

func TestKappaCannoneerRefusesACreatureAndATappedArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elf := pushTapCostPermanent(g, me.ID, "Elf", "Creature — Elf", "G")
	kappa := impKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })
	fillPoolColored(me, "U", 6)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{elf}}); err == nil {
		t.Error("a non-artifact creature was tapped for improvise")
	}
}

func TestPlainSpellCannotImproviseWithoutAGrant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 1)
	bear := impBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Error("a spell with no improvise was improvised")
	}
}

func TestArchwayGrantsImproviseToTheNextSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	archway := pushCatalogPermanent(g, me.ID, "Archway of Innovation", "Land", archwayOracle, false)
	arts := impArtifacts(g, me.ID, 3)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, archway, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating Archway: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := nsPromises(me); len(got) != 1 || !got[0].Improvise {
		t.Fatalf("the promise: %+v", got)
	}

	// An artifact pays the {1} of {1}{G}, but not the {G}.
	bear := impBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[:2]}); err == nil {
		t.Fatal("two artifacts were accepted for {1}{G}")
	}
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[:1]}); err != nil {
		t.Fatalf("improvising the next spell: %v", err)
	}
	if !cgTapped(g, arts[:1]) || cgTapped(g, arts[1:]) {
		t.Error("exactly the named artifact should be tapped")
	}
	if len(nsPromises(me)) != 0 {
		t.Error("the next spell spends the grant")
	}

	// The spell after it has no improvise.
	bear2 := impBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear2) })
	fillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear2.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[1:2]}); err == nil {
		t.Error("the grant covered a second spell")
	}
}

func TestArchwayGrantLapsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	archway := pushCatalogPermanent(g, me.ID, "Archway of Innovation", "Land", archwayOracle, false)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, archway, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating Archway: %v", err)
	}
	passPriorityAroundTable(t, g)
	for g.Turn.ActiveSeat == 0 && len(nsPromises(me)) > 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if len(nsPromises(me)) != 0 {
		t.Error("the improvise grant outlived the turn")
	}
}

func TestKappaCannoneerGrowsAndIsUnblockableWhenAnArtifactEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	arts := impArtifacts(g, me.ID, 5)
	card := impKappa(me.ID)
	kappa := card.InstanceID
	g.WithWriteLock(func() { me.Hand.PushTop(card) })
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, kappa, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, kappa)
	if v.Counters["+1/+1"] != 1 {
		t.Fatalf("entering itself: %d counters, want 1", v.Counters["+1/+1"])
	}

	castTrinket := func(name, typeLine string) {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
		g.WithWriteLock(func() { me.Hand.PushTop(c) })
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
	}
	castTrinket("Trinket", "Artifact")
	v = s58p6View(t, g, kappa)
	if v.Counters["+1/+1"] != 2 {
		t.Errorf("another artifact entered: %d counters, want 2", v.Counters["+1/+1"])
	}
	assertRestrictions(t, g, kappa, game.CantBeBlocked)

	castTrinket("Plain Elf", "Creature — Elf")
	v = s58p6View(t, g, kappa)
	if v.Counters["+1/+1"] != 2 {
		t.Errorf("a non-artifact entered: %d counters, want 2", v.Counters["+1/+1"])
	}
}
