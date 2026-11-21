package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func impBear(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: owner, Controller: owner}
}

func TestKappaCannoneerImprovisePaysGenericButNotBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	arts := effectstest.ImpArtifacts(g, me.ID, 5)
	kappa := effectstest.ImpKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })

	// Five artifacts pay {5}; the {U} still needs mana.
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Fatal("cast with no blue mana succeeded: an artifact must not pay {U}")
	}
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatalf("improvising Kappa Cannoneer: %v", err)
	}
	if !effectstest.CgTapped(g, arts) {
		t.Error("the improvising artifacts were not tapped")
	}
}

func TestKappaCannoneerRefusesMoreArtifactsThanGeneric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	arts := effectstest.ImpArtifacts(g, me.ID, 6)
	kappa := effectstest.ImpKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Error("six artifacts were accepted for five generic")
	}
}

func TestKappaCannoneerRefusesACreatureAndATappedArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	elf := effectstest.PushTapCostPermanent(g, me.ID, "Elf", "Creature — Elf", "G")
	kappa := effectstest.ImpKappa(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(kappa) })
	effectstest.FillPoolColored(me, "U", 6)
	if err := g.CastSpell(me.ID, kappa.InstanceID, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{elf}}); err == nil {
		t.Error("a non-artifact creature was tapped for improvise")
	}
}

func TestPlainSpellCannotImproviseWithoutAGrant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	arts := effectstest.ImpArtifacts(g, me.ID, 1)
	bear := impBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts}); err == nil {
		t.Error("a spell with no improvise was improvised")
	}
}

func TestArchwayGrantsImproviseToTheNextSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	archway := effectstest.PushCatalogPermanent(g, me.ID, "Archway of Innovation", "Land", effectstest.ArchwayOracle, false)
	arts := effectstest.ImpArtifacts(g, me.ID, 3)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, archway, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating Archway: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.NsPromises(me); len(got) != 1 || !got[0].Improvise {
		t.Fatalf("the promise: %+v", got)
	}

	// An artifact pays the {1} of {1}{G}, but not the {G}.
	bear := impBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[:2]}); err == nil {
		t.Fatal("two artifacts were accepted for {1}{G}")
	}
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[:1]}); err != nil {
		t.Fatalf("improvising the next spell: %v", err)
	}
	if !effectstest.CgTapped(g, arts[:1]) || effectstest.CgTapped(g, arts[1:]) {
		t.Error("exactly the named artifact should be tapped")
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Error("the next spell spends the grant")
	}

	// The spell after it has no improvise.
	bear2 := impBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear2) })
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, bear2.InstanceID, game.CastSpellParams{Strict: true, TapIDs: arts[1:2]}); err == nil {
		t.Error("the grant covered a second spell")
	}
}

func TestArchwayGrantLapsAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	archway := effectstest.PushCatalogPermanent(g, me.ID, "Archway of Innovation", "Land", effectstest.ArchwayOracle, false)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, archway, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating Archway: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for g.Turn.ActiveSeat == 0 && len(effectstest.NsPromises(me)) > 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if len(effectstest.NsPromises(me)) != 0 {
		t.Error("the improvise grant outlived the turn")
	}
}

func TestKappaCannoneerGrowsAndIsUnblockableWhenAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	arts := effectstest.ImpArtifacts(g, me.ID, 5)
	card := effectstest.ImpKappa(me.ID)
	kappa := card.InstanceID
	g.WithWriteLock(func() { me.Hand.PushTop(card) })
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, kappa, game.CastSpellParams{Strict: true, TapIDs: arts}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	v := effectstest.S58p6View(t, g, kappa)
	if v.Counters["+1/+1"] != 1 {
		t.Fatalf("entering itself: %d counters, want 1", v.Counters["+1/+1"])
	}

	castTrinket := func(name, typeLine string) {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
		g.WithWriteLock(func() { me.Hand.PushTop(c) })
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	castTrinket("Trinket", "Artifact")
	v = effectstest.S58p6View(t, g, kappa)
	if v.Counters["+1/+1"] != 2 {
		t.Errorf("another artifact entered: %d counters, want 2", v.Counters["+1/+1"])
	}
	effectstest.AssertRestrictions(t, g, kappa, game.CantBeBlocked)

	castTrinket("Plain Elf", "Creature — Elf")
	v = effectstest.S58p6View(t, g, kappa)
	if v.Counters["+1/+1"] != 2 {
		t.Errorf("a non-artifact entered: %d counters, want 2", v.Counters["+1/+1"])
	}
}
