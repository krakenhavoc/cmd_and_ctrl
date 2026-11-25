package testss82m

import (
	"github.com/google/uuid"

	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b226_test.go — Ixalan, batch 226 (#5532): each test plays the real
// catalog entry through the real cast, trigger and combat paths.

const (
	oldB226Castaways = "0432e07d-baf6-4c20-8646-08338c7dc950"
	oldB226Brute     = "675ed28e-4328-4638-aaad-3f47e5c7895f"
	oldB226Pterodon  = "8feb7149-a4dc-491a-a0b4-d2e8b85a97ce"
	oldB226Hammer    = "a690c1de-158a-4f33-a916-ed2a130f92eb"
	oldB226Caller    = "bba6075b-8bbb-4496-80a9-bfa7f003d1e8"
	oldB226Lancer    = "c0018eb9-6aae-4188-8f36-ab5fc69f5903"
	oldB226Brontodon = "fa2d8440-a44a-44fe-8624-0e8b4be53f44"
)

func TestOldB226CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB226Castaways: "Desperate Castaways",
		oldB226Brute:     "Headstrong Brute",
		oldB226Pterodon:  "Pterodon Knight",
		oldB226Hammer:    "Territorial Hammerskull",
		oldB226Caller:    "Kinjalli's Caller",
		oldB226Lancer:    "Imperial Lancer",
		oldB226Brontodon: "Belligerent Brontodon",
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

// --- Desperate Castaways ---

func TestOldB226DesperateCastawaysAttackOnlyWithAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castaways := effectstest.B12Push(g, me.ID, "Desperate Castaways", "Creature — Human Pirate", oldB226Castaways, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(castaways, opp.ID); err == nil {
		t.Fatal("the Castaways attacked with no artifact under their controller's control")
	}
	effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	if err := g.DeclareAttacker(castaways, opp.ID); err == nil {
		t.Fatal("an artifact an opponent controls let the Castaways attack")
	}
	effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	if err := g.DeclareAttacker(castaways, opp.ID); err != nil {
		t.Errorf("the Castaways could not attack with an artifact: %v", err)
	}
}

// --- Headstrong Brute ---

func TestOldB226HeadstrongBruteCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	brute := effectstest.B12Push(g, opp.ID, "Headstrong Brute", "Creature — Orc Pirate", oldB226Brute, 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(brute, attacker); err == nil {
		t.Error("the Brute blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}
}

func TestOldB226HeadstrongBruteHasMenaceOnlyWithAnotherPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := effectstest.B12Push(g, me.ID, "Headstrong Brute", "Creature — Orc Pirate", oldB226Brute, 3, 3)
	if effectstest.HasKeywordOf(g, brute, "menace") {
		t.Fatal("the Brute has menace as the only Pirate (it is not 'another')")
	}
	effectstest.B12Creature(g, opp.ID, "Their Pirate", "Creature — Human Pirate", 1, 1)
	if effectstest.HasKeywordOf(g, brute, "menace") {
		t.Fatal("an opponent's Pirate gave the Brute menace")
	}
	pirate := effectstest.B12Creature(g, me.ID, "My Pirate", "Creature — Human Pirate", 1, 1)
	if !effectstest.HasKeywordOf(g, brute, "menace") {
		t.Error("the Brute lacks menace with another Pirate")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pirate) })
	if effectstest.HasKeywordOf(g, brute, "menace") {
		t.Error("the Brute kept menace after the other Pirate left")
	}
}

// --- Pterodon Knight and Imperial Lancer ---

func TestOldB226KnightsGainTheirKeywordWhileYouControlADinosaur(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle, keyword string
		power, toughness                int
	}{
		{"Pterodon Knight", "Creature — Human Knight", oldB226Pterodon, "flying", 3, 3},
		{"Imperial Lancer", "Creature — Human Knight", oldB226Lancer, "double strike", 1, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		knight := effectstest.B12Push(g, me.ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		if effectstest.HasKeywordOf(g, knight, c.keyword) {
			t.Fatalf("%s has %s with no Dinosaur", c.name, c.keyword)
		}
		effectstest.B12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 2, 2)
		if effectstest.HasKeywordOf(g, knight, c.keyword) {
			t.Fatalf("%s has %s off an opponent's Dinosaur", c.name, c.keyword)
		}
		dino := effectstest.B12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 2, 2)
		if !effectstest.HasKeywordOf(g, knight, c.keyword) {
			t.Errorf("%s lacks %s with a Dinosaur", c.name, c.keyword)
		}
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dino) })
		if effectstest.HasKeywordOf(g, knight, c.keyword) {
			t.Errorf("%s kept %s after the Dinosaur left", c.name, c.keyword)
		}
	}
}

// --- Territorial Hammerskull ---

func TestOldB226TerritorialHammerskullTapsTargetCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hammer := effectstest.B12Push(g, me.ID, "Territorial Hammerskull", "Creature — Dinosaur", oldB226Hammer, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hammer)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Fatal("a creature I control was an acceptable target")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, theirs) {
		t.Error("the opposing creature was not tapped")
	}
	if effectstest.IsTapped(g, mine) {
		t.Error("my own creature was tapped")
	}
}

// --- Kinjalli's Caller ---

func TestOldB226KinjallisCallerDiscountsDinosaurSpellsOnce(t *testing.T) {
	cases := []struct {
		typeLine, cost string
		lands          int
		ok             bool
	}{
		{"Creature — Dinosaur", "{2}{R}", 2, true},
		{"Creature — Dinosaur", "{2}{R}", 1, false},
		{"Creature — Dinosaur", "{R}", 1, true}, // the coloured cost is never discounted
		{"Creature — Human Warrior", "{2}{R}", 2, false},
	}
	for _, tc := range cases {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Kinjalli's Caller", "Creature — Human Cleric", oldB226Caller, 0, 3)
		_, err := effectstest.TrekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s %s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.cost, tc.lands, err, tc.ok)
		}
	}
}

// --- Belligerent Brontodon ---

func TestOldB226BelligerentBrontodonAssignsDamageByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saur := effectstest.B12Push(g, me.ID, "Belligerent Brontodon", "Creature — Dinosaur", oldB226Brontodon, 4, 6)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 5)
	for _, id := range []uuid.UUID{saur, wall} {
		c := effectstest.LayeredCard(t, g, id)
		if !c.Effective().DamageByToughnessAlways {
			t.Errorf("%s does not assign damage by toughness", c.Name)
		}
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.Effective().DamageByToughnessAlways {
		t.Error("an opponent's creature assigns damage by toughness")
	}
	life := opp.Life
	effectstest.BrAttack(t, g, wall)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if opp.Life != life-5 {
		t.Errorf("the 0/5 wall dealt %d, want 5 (its toughness)", life-opp.Life)
	}
}
