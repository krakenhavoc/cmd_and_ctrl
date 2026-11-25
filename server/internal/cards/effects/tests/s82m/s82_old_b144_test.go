package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b144_test.go — Throne of Eldraine, S82 older-sets batch 144
// (#5516): +1/+1 and other counters. Each test plays the real catalog entry.

const (
	oldB144YorvoOracle      = "00a15abc-bdac-4977-90a0-f08d41941d01"
	oldB144NobleOracle      = "4c23ae3a-108f-4a09-b89f-8c43c33dd144"
	oldB144WandermareOracle = "7c25470c-7be0-495c-b756-6a4fc230d660"
	oldB144MirrorOracle     = "9acf2701-2ba1-46fa-ad4a-b130a58181d3"
	oldB144FaerieOracle     = "db6fbdc9-1c45-416f-a08e-2082f65adb65"
	oldB144PathlightOracle  = "fe87e290-8e9d-4269-ad1f-bb31e6eeb7da"
)

func TestOldB144EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB144YorvoOracle: "Yorvo, Lord of Garenbrig", oldB144NobleOracle: "Malevolent Noble",
		oldB144WandermareOracle: "Wandermare", oldB144MirrorOracle: "The Magic Mirror",
		oldB144FaerieOracle: "Animating Faerie", oldB144FaerieOracle + "#1": "Bring to Life",
		oldB144PathlightOracle: "Mysterious Pathlighter",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Yorvo, Lord of Garenbrig ---------------------------------------------

func oldB144CastCreature(t *testing.T, g *game.Game, me *game.Player, name, colors string, power int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: power, Owner: me.ID, Controller: me.ID}
	if colors != "" {
		c.Colors = []string{colors}
	}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB144YorvoEntersWithFourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastCatalogSpell(t, g, "Yorvo, Lord of Garenbrig", "Legendary Creature — Giant Noble", oldB144YorvoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, id); got != 4 {
		t.Fatalf("Yorvo entered with %d counters, want 4", got)
	}
}

func TestOldB144YorvoGrowsForAGreenCreatureAndAgainWhenItIsBigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	yorvo := effectstest.B12Push(g, me.ID, "Yorvo, Lord of Garenbrig", "Legendary Creature — Giant Noble", oldB144YorvoOracle, 0, 0)
	effectstest.PutCounters(t, g, yorvo, game.CounterPlusOne, 4)
	oldB144CastCreature(t, g, me, "Small Elf", "G", 2)
	if got := effectstest.PlusOneCounters(g, yorvo); got != 5 {
		t.Fatalf("after a smaller green creature: %d counters, want 5", got)
	}
	oldB144CastCreature(t, g, me, "Big Beast", "G", 7)
	if got := effectstest.PlusOneCounters(g, yorvo); got != 7 {
		t.Fatalf("after a green 7/7 (greater than Yorvo's 6... then one counter and another): %d counters, want 7", got)
	}
}

func TestOldB144YorvoIgnoresNonGreenAndOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	yorvo := effectstest.B12Push(g, me.ID, "Yorvo, Lord of Garenbrig", "Legendary Creature — Giant Noble", oldB144YorvoOracle, 0, 0)
	effectstest.PutCounters(t, g, yorvo, game.CounterPlusOne, 4)
	oldB144CastCreature(t, g, me, "Red Goblin", "R", 5)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Elf", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, yorvo); got != 4 {
		t.Fatalf("a non-green or opposing creature grew Yorvo to %d counters, want 4", got)
	}
}

// --- Malevolent Noble -------------------------------------------------------

func TestOldB144MalevolentNobleSacrificesAnArtifactOrAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	noble := effectstest.B12Push(g, me.ID, "Malevolent Noble", "Creature — Human Noble", oldB144NobleOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
		Owner: me.ID, Controller: me.ID})

	activate := func(sac uuid.UUID) error {
		effectstest.B10AddMana(me, "B", "B")
		return g.ActivateCatalogAbility(me.ID, noble, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sac}})
	}
	if err := activate(noble); err == nil {
		t.Fatal("the Noble sacrificed itself: it says another creature")
	}
	if err := activate(land); err == nil {
		t.Fatal("a land was sacrificed")
	}
	if err := activate(fodder); err != nil {
		t.Fatalf("sacrificing another creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, noble); got != 1 {
		t.Fatalf("after one sacrifice: %d counters, want 1", got)
	}
	if err := activate(rock); err != nil {
		t.Fatalf("sacrificing an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, noble); got != 2 {
		t.Fatalf("after two sacrifices: %d counters, want 2", got)
	}
}

// --- Wandermare ---------------------------------------------------------------

func TestOldB144WandermareCountsACreatureSpellWithAnAdventureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mare := effectstest.B12Push(g, me.ID, "Wandermare", "Creature — Horse", oldB144WandermareOracle, 3, 3)
	effectstest.B13PlayAs(t, g, 0, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, mare); got != 0 {
		t.Fatalf("a plain creature spell gave %d counters", got)
	}
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)
	if got := effectstest.PlusOneCounters(g, mare); got != 0 {
		t.Fatalf("the Adventure half gave %d counters", got)
	}
	_, id2 := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id2, 0, nil)
	if got := effectstest.PlusOneCounters(g, mare); got != 1 {
		t.Fatalf("casting the creature half gave %d counters, want 1", got)
	}
}

// --- Mysterious Pathlighter ----------------------------------------------------

func TestOldB144PathlighterGivesAnAdventureCreatureAnExtraCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pl := effectstest.B12Push(g, me.ID, "Mysterious Pathlighter", "Creature — Faerie", oldB144PathlightOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, pl, "flying") {
		t.Error("no flying")
	}
	bear := oldB144CastCreature(t, g, me, "Bear", "G", 2)
	if got := effectstest.PlusOneCounters(g, bear); got != 0 {
		t.Fatalf("a creature with no Adventure entered with %d counters", got)
	}
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if got := effectstest.PlusOneCounters(g, id); got != 1 {
		t.Fatalf("an Adventure creature entered with %d counters, want 1", got)
	}
	if c := effectstest.B12Card(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("P/T %d/%d, want 2/2 (a 1/1 with the counter)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB144PathlighterDoesNotBoostAnOpponentsAdventureCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, g.Seats[1].ID, "Mysterious Pathlighter", "Creature — Faerie", oldB144PathlightOracle, 2, 2)
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if got := effectstest.PlusOneCounters(g, id); got != 0 {
		t.Fatalf("an opponent's Pathlighter gave %d counters", got)
	}
}

// --- Animating Faerie // Bring to Life -----------------------------------------

func oldB144FaerieInHand(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	return effectstest.HobA6HandAdventure(t, g, oldB144FaerieOracle,
		effectstest.HobA6Face("Animating Faerie", "Creature — Faerie", "{2}{U}", 2, 2, "U"),
		effectstest.HobA6Face("Bring to Life", "Sorcery — Adventure", "{2}{U}", 0, 0, "U"))
}

func TestOldB144BringToLifeMakesANoncreatureArtifactAFourFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID})
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	_, id := oldB144FaerieInHand(t, g)
	for _, bad := range []uuid.UUID{theirs, golem} {
		effectstest.B10AddMana(me, "U", "U", "U")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1, Strict: true, Targets: effectstest.CardTarget(bad)}); err == nil {
			t.Fatal("an illegal target was accepted")
		}
	}
	effectstest.B10AddMana(me, "U", "U", "U")
	effectstest.HobA6MustCast(t, g, me, id, 1, effectstest.CardTarget(rock))
	if !effectstest.HasEffectiveType(t, g, rock, "Creature") || !effectstest.HasEffectiveType(t, g, rock, "Artifact") {
		t.Fatalf("the artifact is not an artifact creature: %v", effectstest.EffectiveTypes(t, g, rock))
	}
	if c := effectstest.B12Card(t, g, rock); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("P/T %d/%d, want 4/4 (0/0 with four counters)", c.CurrentPower(), c.CurrentToughness())
	}
	if got := effectstest.PlusOneCounters(g, rock); got != 4 {
		t.Errorf("%d counters, want 4", got)
	}
	// Adventure: the card went to exile and the creature can be cast later.
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestOldB144AnimatingFaerieIsAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.B12Push(g, me.ID, "Animating Faerie", "Creature — Faerie", oldB144FaerieOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, f, "flying") {
		t.Error("no flying")
	}
}

// --- The Magic Mirror -------------------------------------------------------------

func oldB144Mirror() game.Card {
	return game.Card{Name: "The Magic Mirror", TypeLine: "Legendary Artifact", ManaCost: "{6}{U}{U}{U}", OracleID: oldB144MirrorOracle}
}

func TestOldB144MagicMirrorCostsOneLessPerInstantAndSorceryInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.SocB19CastWithPool(g, me, oldB144Mirror(), "U", 8, game.CastSpellParams{}); err == nil {
		t.Fatal("cast for 8 with an empty graveyard")
	}
	effectstest.PushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	effectstest.PushGraveyardCardWithTypeLineID(me, "Ritual", "Sorcery")
	effectstest.PushGraveyardCardWithTypeLineID(me, "Bear", "Creature — Bear")
	if err := effectstest.SocB19CastWithPool(g, me, oldB144Mirror(), "U", 6, game.CastSpellParams{}); err == nil {
		t.Fatal("cast for 6 with two instant/sorcery cards (cost 7): a creature card discounted it")
	}
	if err := effectstest.SocB19CastWithPool(g, me, oldB144Mirror(), "U", 7, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for 7 with two instant/sorcery cards: %v", err)
	}
}

func TestOldB144MagicMirrorGrowsItsKnowledgeAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mirror := effectstest.PushCatalogPermanent(g, me.ID, "The Magic Mirror", "Legendary Artifact", oldB144MirrorOracle, false)
	if spec, _ := effects.Lookup(oldB144MirrorOracle); !spec.NoMaxHandSize {
		t.Error("the Mirror does not declare no maximum hand size")
	}
	upkeep := func() int {
		for g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		before := effectstest.HandCount(g, me)
		effectstest.PassPriorityAroundTable(t, g)
		return effectstest.HandCount(g, me) - before
	}
	if got := upkeep(); got != 1 {
		t.Fatalf("first upkeep drew %d, want 1", got)
	}
	if got := effectstest.CounterCountOf(t, g, mirror, "knowledge"); got != 1 {
		t.Fatalf("%d knowledge counters, want 1", got)
	}
	if got := upkeep(); got != 2 {
		t.Fatalf("second upkeep drew %d, want 2", got)
	}
	if got := effectstest.CounterCountOf(t, g, mirror, "knowledge"); got != 2 {
		t.Fatalf("%d knowledge counters, want 2", got)
	}
}
