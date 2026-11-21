package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b16_cards_test.go — Innistrad Remastered sorceries, GAP batch 16
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapSpiderSpawningOracle     = "da113cc9-bb2b-4af4-9c43-32c165b87363"
	gapTraverseOracle           = "c9dd2bf5-32e0-47f7-9a49-501c262d745b"
	gapRiseFromTheTidesOracle   = "390b862a-2a85-44bd-832c-f5bbd3eb4ea0"
	gapAngelicPurgeOracle       = "b262a057-9ace-48b4-b418-1bb751ccb300"
	gapGatherTheTownsfolkOracle = "ff45677d-977e-4ef5-859c-52393ed71a6c"
)

func TestGapB16EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapSpiderSpawningOracle:     "Spider Spawning",
		gapTraverseOracle:           "Traverse the Ulvenwald",
		gapRiseFromTheTidesOracle:   "Rise from the Tides",
		gapAngelicPurgeOracle:       "Angelic Purge",
		gapGatherTheTownsfolkOracle: "Gather the Townsfolk",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for _, oracle := range []string{gapSpiderSpawningOracle} {
		if spec, _ := Lookup(oracle); len(spec.AlternativeCosts) != 1 {
			t.Errorf("%s has no flashback", spec.Name)
		}
	}
}

func TestGapB16SpiderSpawningMakesOneSpiderPerCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	pushCatalogGraveyardCard(me, "Dead Elf", "Creature — Elf", "", 1, 1)
	pushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	castCatalogSpell(t, g, "Spider Spawning", "Sorcery", gapSpiderSpawningOracle, nil)
	passPriorityAroundTable(t, g)
	spiders := battlefieldIDsNamed(g, "Spider")
	if len(spiders) != 2 {
		t.Fatalf("%d Spiders, want 2 (one per creature card)", len(spiders))
	}
	assertKeywords(t, g, spiders[0], "reach")
	if effectivePower(t, g, spiders[0]) != 1 || effectiveToughness(t, g, spiders[0]) != 2 {
		t.Error("a Spider is not 1/2")
	}
}

func TestGapB16SpiderSpawningWithNoCreatureCardsMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Spider Spawning", "Sorcery", gapSpiderSpawningOracle, nil)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Spider")); n != 0 {
		t.Errorf("%d Spiders from an empty graveyard", n)
	}
}

func TestGapB16TraverseTheUlvenwaldFetchesABasicLandThenACreatureOrLandWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	castCatalogSpell(t, g, "Traverse the Ulvenwald", "Sorcery", gapTraverseOracle, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !b02bHandHasNamed(me, "Forest") {
		t.Fatal("one basic land in the library: it goes straight to hand")
	}
	if b02bHandHasNamed(me, "Bear") || b02bHandHasNamed(me, "Sol Ring") {
		t.Fatal("a nonland card was fetched without delirium")
	}

	for _, c := range []game.Card{
		{Name: "Dead Bear", TypeLine: "Creature — Bear"},
		{Name: "Dead Rock", TypeLine: "Artifact"},
		{Name: "Dead Bolt", TypeLine: "Instant"},
		{Name: "Dead Land", TypeLine: "Land"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	castCatalogSpell(t, g, "Traverse the Ulvenwald", "Sorcery", gapTraverseOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("with delirium the searcher chooses between a creature and a land")
	}
	if searchOptionNamed(g, c, "Sol Ring") != uuid.Nil {
		t.Error("an artifact is offered: it is a creature or land search")
	}
	answerSearchNamed(t, g, me.ID, "Bear")
	if !b02bHandHasNamed(me, "Bear") {
		t.Error("the chosen creature card did not reach the hand")
	}
}

func TestGapB16RiseFromTheTidesMakesTappedZombiesPerInstantAndSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	pushCatalogGraveyardCard(me, "Dead Wrath", "Sorcery", "", 0, 0)
	pushCatalogGraveyardCard(me, "Dead Arcane", "Instant — Arcane", "", 0, 0)
	pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Rise from the Tides", "Sorcery", gapRiseFromTheTidesOracle, nil)
	passPriorityAroundTable(t, g)
	zombies := battlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 3 {
		t.Fatalf("%d Zombies, want 3", len(zombies))
	}
	for _, z := range zombies {
		if !tappedForTest(t, g, z) {
			t.Error("a Zombie entered untapped")
		}
		if effectivePower(t, g, z) != 2 || effectiveToughness(t, g, z) != 2 {
			t.Error("a Zombie is not 2/2")
		}
	}
}

func TestGapB16AngelicPurgeSacrificesAPermanentAndExilesTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Angelic Purge", TypeLine: "Sorcery", OracleID: gapAngelicPurgeOracle,
		ManaCost: "{2}{W}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "W", "W", "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(victim), Strict: true}); err == nil {
		t.Fatal("cast without sacrificing a permanent")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(land), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed permanent is still out")
	}
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the target was not exiled")
	}
}

func TestGapB16GatherTheTownsfolkMakesTwoOrFiveByLife(t *testing.T) {
	for _, tc := range []struct{ life, want int }{{20, 2}, {6, 2}, {5, 5}, {1, 5}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.Life = tc.life })
		castCatalogSpell(t, g, "Gather the Townsfolk", "Sorcery", gapGatherTheTownsfolkOracle, nil)
		passPriorityAroundTable(t, g)
		humans := battlefieldIDsNamed(g, "Human")
		if len(humans) != tc.want {
			t.Errorf("life %d: %d Human tokens, want %d", tc.life, len(humans), tc.want)
			continue
		}
		if effectivePower(t, g, humans[0]) != 1 || controllerOf(t, g, humans[0]) != me.ID {
			t.Errorf("life %d: want 1/1 Humans under your control", tc.life)
		}
	}
}
