package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for _, oracle := range []string{gapSpiderSpawningOracle} {
		if spec, _ := effects.Lookup(oracle); len(spec.AlternativeCosts) != 1 {
			t.Errorf("%s has no flashback", spec.Name)
		}
	}
}

func TestGapB16SpiderSpawningMakesOneSpiderPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(me, "Dead Elf", "Creature — Elf", "", 1, 1)
	effectstest.PushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Spider Spawning", "Sorcery", gapSpiderSpawningOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	spiders := effectstest.BattlefieldIDsNamed(g, "Spider")
	if len(spiders) != 2 {
		t.Fatalf("%d Spiders, want 2 (one per creature card)", len(spiders))
	}
	effectstest.AssertKeywords(t, g, spiders[0], "reach")
	if effectstest.EffectivePower(t, g, spiders[0]) != 1 || effectstest.EffectiveToughness(t, g, spiders[0]) != 2 {
		t.Error("a Spider is not 1/2")
	}
}

func TestGapB16SpiderSpawningWithNoCreatureCardsMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Spider Spawning", "Sorcery", gapSpiderSpawningOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Spider")); n != 0 {
		t.Errorf("%d Spiders from an empty graveyard", n)
	}
}

func TestGapB16TraverseTheUlvenwaldFetchesABasicLandThenACreatureOrLandWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	effectstest.CastCatalogSpell(t, g, "Traverse the Ulvenwald", "Sorcery", gapTraverseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !effectstest.B02bHandHasNamed(me, "Forest") {
		t.Fatal("one basic land in the library: it goes straight to hand")
	}
	if effectstest.B02bHandHasNamed(me, "Bear") || effectstest.B02bHandHasNamed(me, "Sol Ring") {
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
	effectstest.CastCatalogSpell(t, g, "Traverse the Ulvenwald", "Sorcery", gapTraverseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("with delirium the searcher chooses between a creature and a land")
	}
	if effectstest.SearchOptionNamed(g, c, "Sol Ring") != uuid.Nil {
		t.Error("an artifact is offered: it is a creature or land search")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Bear")
	if !effectstest.B02bHandHasNamed(me, "Bear") {
		t.Error("the chosen creature card did not reach the hand")
	}
}

func TestGapB16RiseFromTheTidesMakesTappedZombiesPerInstantAndSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	effectstest.PushCatalogGraveyardCard(me, "Dead Wrath", "Sorcery", "", 0, 0)
	effectstest.PushCatalogGraveyardCard(me, "Dead Arcane", "Instant — Arcane", "", 0, 0)
	effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rise from the Tides", "Sorcery", gapRiseFromTheTidesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	zombies := effectstest.BattlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 3 {
		t.Fatalf("%d Zombies, want 3", len(zombies))
	}
	for _, z := range zombies {
		if !effectstest.TappedForTest(t, g, z) {
			t.Error("a Zombie entered untapped")
		}
		if effectstest.EffectivePower(t, g, z) != 2 || effectstest.EffectiveToughness(t, g, z) != 2 {
			t.Error("a Zombie is not 2/2")
		}
	}
}

func TestGapB16AngelicPurgeSacrificesAPermanentAndExilesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Permanent(g, me.ID, "Fodder", "Artifact")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Angelic Purge", TypeLine: "Sorcery", OracleID: gapAngelicPurgeOracle,
		ManaCost: "{2}{W}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W", "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(victim), Strict: true}); err == nil {
		t.Fatal("cast without sacrificing a permanent")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(land), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(victim), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed permanent is still out")
	}
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the target was not exiled")
	}
}

func TestGapB16GatherTheTownsfolkMakesTwoOrFiveByLife(t *testing.T) {
	for _, tc := range []struct{ life, want int }{{20, 2}, {6, 2}, {5, 5}, {1, 5}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.Life = tc.life })
		effectstest.CastCatalogSpell(t, g, "Gather the Townsfolk", "Sorcery", gapGatherTheTownsfolkOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		humans := effectstest.BattlefieldIDsNamed(g, "Human")
		if len(humans) != tc.want {
			t.Errorf("life %d: %d Human tokens, want %d", tc.life, len(humans), tc.want)
			continue
		}
		if effectstest.EffectivePower(t, g, humans[0]) != 1 || effectstest.ControllerOf(t, g, humans[0]) != me.ID {
			t.Errorf("life %d: want 1/1 Humans under your control", tc.life)
		}
	}
}
