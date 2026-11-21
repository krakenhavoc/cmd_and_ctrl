package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b19_cards_test.go — Secrets of Strixhaven (sos), first batch (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB19ManawurmOracle     = "a0657586-0404-4bb0-a92d-25f040d560c1"
	socB19AbigaleOracle      = "2f5f46ed-b8aa-4864-bd20-17281d4632bf"
	socB19PaintmageOracle    = "bc899d5c-0a91-4f19-a3e4-3078ae966110"
	socB19EvolutionOracle    = "ce1a6c1f-6b4e-4d19-b256-48d073b359b7"
	socB19EaterOracle        = "02125db4-4507-467a-ac0c-406de2c7d533"
	socB19ResponseOracle     = "bf7ec320-64bb-46ca-aac0-ca3d4224b69e"
	socB19AugmenterOracle    = "d87f0e52-486e-49b4-9cf1-12e238857340"
	socB19AngerOracle        = "e0828e8d-f01f-4088-9123-6d923ddb3242"
	socB19AntiquitiesOracle  = "9c45682f-48b5-4e58-9277-ff4211f389ae"
	socB19GeometryOracle     = "88fc179c-055b-47c2-b830-1fd2b1377147"
	socB19PlainCreatureTypes = "Creature — Bear"
)

func TestSocB19EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB19ManawurmOracle:       "Aberrant Manawurm",
		socB19AbigaleOracle:        "Abigale, Poet Laureate",
		socB19AbigaleOracle + "#1": "Heroic Stanza",
		socB19PaintmageOracle:      "Abstract Paintmage",
		socB19EvolutionOracle:      "Additive Evolution",
		socB19EaterOracle:          "Adventurous Eater",
		socB19EaterOracle + "#1":   "Have a Bite",
		socB19ResponseOracle:       "Ajani's Response",
		socB19AugmenterOracle:      "Ambitious Augmenter",
		socB19AngerOracle:          "Ancestral Anger",
		socB19AntiquitiesOracle:    "Antiquities on the Loose",
		socB19GeometryOracle:       "Applied Geometry",
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
}

// --- Aberrant Manawurm ----------------------------------------------------

func TestSocB19ManawurmGetsPlusXForTheManaSpentOnAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worm := effectstest.B12Push(g, me.ID, "Aberrant Manawurm", "Creature — Wurm", socB19ManawurmOracle, 2, 5)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{G}"}, "G", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.SocB14PT(t, g, worm); p != 5 {
		t.Errorf("power %d after a three-mana sorcery, want 2+3", p)
	}
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "One", TypeLine: "Instant", ManaCost: "{G}"}, "G", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.SocB14PT(t, g, worm); p != 6 {
		t.Errorf("power %d after a second one-mana instant, want 6", p)
	}
	// Until end of turn only.
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.SocB14PT(t, g, worm); p != 2 {
		t.Errorf("power %d on the next turn, want it back at 2", p)
	}
}

func TestSocB19ManawurmIgnoresACreatureSpellAndAFreeSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	worm := effectstest.B12Push(g, me.ID, "Aberrant Manawurm", "Creature — Wurm", socB19ManawurmOracle, 2, 5)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: socB19PlainCreatureTypes, ManaCost: "{1}{G}", Power: 2, Toughness: 2}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.SocB14PT(t, g, worm); p != 2 {
		t.Errorf("power %d after a creature spell, want 2", p)
	}
	// A spell cast without paying spent no mana: +0/+0.
	effectstest.SocB4Spell(t, g, "Free", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.SocB14PT(t, g, worm); p != 2 {
		t.Errorf("power %d after a free instant, want 2", p)
	}
}

// --- Abigale, Poet Laureate // Heroic Stanza ------------------------------

func socB19CastAbigale(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB19AbigaleOracle, []string{"flying"}, "Abigale, Poet Laureate", "Legendary Creature — Bird Bard", "{1}{W}{B}", 2, 3,
		"Heroic Stanza", "Sorcery", "{1}{W/B}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func TestSocB19AbigalePreparesWhenYouCastACreatureSpellAndStanzaPutsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	abigale := socB19CastAbigale(t, g, me)
	if g.IsPreparedForEffect(abigale) {
		t.Fatal("prepared on entering")
	}
	// A noncreature spell does not prepare her.
	effectstest.SocB4Spell(t, g, "Bolt", "Instant", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(abigale) {
		t.Fatal("prepared by an instant")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	effectstest.SocB4Spell(t, g, "Elk", "Creature — Elk", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(abigale) {
		t.Fatal("not prepared after casting a creature spell")
	}
	cp, ok := effectstest.PrepareCopyOf(g, "Heroic Stanza")
	if !ok {
		t.Fatal("no Heroic Stanza copy in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast Heroic Stanza: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the target %d, want 1", n)
	}
	if g.IsPreparedForEffect(abigale) {
		t.Error("still prepared after casting the copy")
	}
}

// --- Abstract Paintmage ---------------------------------------------------

func TestSocB19PaintmageAddsURForInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Abstract Paintmage", "Creature — Djinn Sorcerer", socB19PaintmageOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.ManaPool.Total(); n != 2 {
		t.Fatalf("pool %d after the first main phase began, want {U}{R}", n)
	}
	creature := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Ox", TypeLine: "Creature — Ox", ManaCost: "{U}{R}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the restricted mana paid for a creature spell")
	}
	sorcery := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Spark", TypeLine: "Sorcery", ManaCost: "{U}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the restricted mana did not pay for a sorcery: %v", err)
	}
}

func TestSocB19PaintmageDoesNotTriggerOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pm := effectstest.B12Push(g, me.ID, "Abstract Paintmage", "Creature — Djinn Sorcerer", socB19PaintmageOracle, 2, 2)
	effectstest.EndTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if n := effectstest.TriggersOnStackFrom(g, pm); n != 0 {
		t.Errorf("%d Paintmage triggers on an opponent's main phase", n)
	}
}

// --- Additive Evolution ---------------------------------------------------

func TestSocB19AdditiveEvolutionMakesAThreeCounterFractalAndGrowsACreatureEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Additive Evolution", "Enchantment", socB19EvolutionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	fractals := socB4Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("Fractals %d, want 1", len(fractals))
	}
	if p, tg := effectstest.SocB14PT(t, g, fractals[0].InstanceID); p != 3 || tg != 3 {
		t.Errorf("Fractal is %d/%d, want 3/3", p, tg)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("Bear is %d/%d after the combat trigger, want 3/3", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the target did not gain vigilance")
	}
}

func TestSocB19AdditiveEvolutionOnlyTriggersOnYourOwnCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Additive Evolution", "Enchantment", socB19EvolutionOracle, false)
	effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	effectstest.EndTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if n := effectstest.TriggersOnStackFrom(g, id); n != 0 {
		t.Errorf("%d triggers at an opponent's beginning of combat", n)
	}
}

// --- Adventurous Eater // Have a Bite -------------------------------------

func TestSocB19AdventurousEaterEntersPreparedAndHaveABiteCountersAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB19EaterOracle, nil, "Adventurous Eater", "Creature — Human Warlock", "{2}{B}", 3, 2,
		"Have a Bite", "Sorcery", "{B}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("did not enter prepared")
	}
	cp, ok := effectstest.PrepareCopyOf(g, "Have a Bite")
	if !ok {
		t.Fatal("no Have a Bite copy in exile")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	life := me.Life
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast Have a Bite: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the target %d, want 1", n)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

// --- Ajani's Response -----------------------------------------------------

func TestSocB19AjanisResponseCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { effectstest.B16Tap(g, tapped) })
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Ajani's Response", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: socB19ResponseOracle}

	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: effectstest.CardRefs(untapped)}); err == nil {
		t.Fatal("cast at the full price with two mana against an untapped creature")
	}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: effectstest.CardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature died")
	}
}

func TestSocB19AjanisResponseAtFullPriceDestroysAnUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Ajani's Response", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: socB19ResponseOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 5, game.CastSpellParams{Targets: effectstest.CardRefs(ox)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ox) {
		t.Error("the creature survived")
	}
}

// --- Ambitious Augmenter --------------------------------------------------

func socB19Augmenter(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Ambitious Augmenter", "Creature — Turtle Wizard", socB19AugmenterOracle, 1, 1)
}

func TestSocB19AugmenterIncrementsOnlyWhenTheManaSpentExceedsPowerOrToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aug := socB19Augmenter(g, me)
	effectstest.AdvanceToMain(t, g)
	cast := func(cost string, n int) {
		t.Helper()
		if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: cost}, "G", n, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("{G}", 1) // 1 is not greater than 1/1
	if n := effectstest.CountersOn(g, aug, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters %d after a one-mana spell, want 0", n)
	}
	cast("{1}{G}", 2) // 2 > 1: now 2/2
	if n := effectstest.CountersOn(g, aug, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a two-mana spell, want 1", n)
	}
	cast("{1}{G}", 2) // 2 is not greater than 2/2
	if n := effectstest.CountersOn(g, aug, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a second two-mana spell, want still 1", n)
	}
	cast("{2}{G}", 3) // 3 > 2: now 3/3
	if n := effectstest.CountersOn(g, aug, game.CounterPlusOne); n != 2 {
		t.Fatalf("counters %d after a three-mana spell, want 2", n)
	}
}

func TestSocB19AugmenterRechecksTheConditionAsTheTriggerResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aug := socB19Augmenter(g, me)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}{G}"}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := effectstest.TriggersOnStackFrom(g, aug); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	// In response it grows to 2/2: two mana is no longer greater (CR 603.4).
	effectstest.AddCountersForTest(g, aug, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, aug, game.CounterPlusOne); n != 1 {
		t.Errorf("counters %d, want the one added in response and no more", n)
	}
}

func TestSocB19AugmenterDiesWithCountersAndTheFractalTakesThemAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aug := socB19Augmenter(g, me)
	effectstest.AddCountersForTest(g, aug, game.CounterPlusOne, 2)
	effectstest.AddCountersForTest(g, aug, "stun", 1)
	_ = opp
	effectstest.CastCatalogSpell(t, g, "Ajani's Response", "Instant", socB19ResponseOracle, effectstest.CardRefs(aug))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aug) {
		t.Fatal("the Augmenter survived")
	}
	fractals := socB4Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("Fractals %d, want 1", len(fractals))
	}
	id := fractals[0].InstanceID
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters on the Fractal %d, want 2", n)
	}
	if n := effectstest.CountersOn(g, id, "stun"); n != 1 {
		t.Errorf("stun counters on the Fractal %d, want 1", n)
	}
	if p, tg := effectstest.SocB14PT(t, g, id); p != 2 || tg != 2 {
		t.Errorf("Fractal is %d/%d, want 2/2", p, tg)
	}
}

func TestSocB19AugmenterDiesWithoutCountersMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aug := socB19Augmenter(g, me)
	effectstest.CastCatalogSpell(t, g, "Ajani's Response", "Instant", socB19ResponseOracle, effectstest.CardRefs(aug))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aug) {
		t.Fatal("the Augmenter survived")
	}
	if n := len(socB4Fractals(g, me.ID)); n != 0 {
		t.Errorf("Fractals %d, want none for a creature that died with no counters", n)
	}
}

// --- Ancestral Anger ------------------------------------------------------

func TestSocB19AncestralAngerCountsCardsNamedAncestralAngerInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ancestral Anger", TypeLine: "Sorcery", OracleID: socB19AngerOracle, Owner: me.ID, Controller: me.ID})
	}
	// Another card in the graveyard does not count.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Other", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	card := game.Card{Name: "Ancestral Anger", TypeLine: "Sorcery", ManaCost: "{R}", OracleID: socB19AngerOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "R", 1, game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 5 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 5/2 (X = 1 + 2)", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the spell draws a card", me.Hand.Size(), hand+1)
	}
	// Until end of turn only.
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.SocB14PT(t, g, bear); p != 2 {
		t.Errorf("power %d next turn, want 2", p)
	}
}

func TestSocB19AncestralAngerAloneIsPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ancestral Anger", "Sorcery", socB19AngerOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.SocB14PT(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3 (X = 1)", p)
	}
}

// --- Antiquities on the Loose ---------------------------------------------

func TestSocB19AntiquitiesFromHandMakesTwoPlainSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Antiquities on the Loose", TypeLine: "Sorcery", ManaCost: "{1}{W}{W}", OracleID: socB19AntiquitiesOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.SocB19Named(g, me.ID, "Spirit")
	if len(spirits) != 2 {
		t.Fatalf("Spirits %d, want 2", len(spirits))
	}
	for _, s := range spirits {
		if p, tg := effectstest.SocB14PT(t, g, s.InstanceID); p != 2 || tg != 2 {
			t.Errorf("Spirit is %d/%d, want 2/2 with no counter from a hand cast", p, tg)
		}
	}
}

func TestSocB19AntiquitiesFlashbackPutsACounterOnEachSpiritYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	older := effectstest.B12Creature(g, me.ID, "Old Spirit", "Creature — Spirit", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Antiquities on the Loose", TypeLine: "Sorcery", ManaCost: "{1}{W}{W}",
		OracleID: socB19AntiquitiesOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "W", 6)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.SocB19Named(g, me.ID, "Spirit")
	if len(spirits) != 2 {
		t.Fatalf("token Spirits %d, want 2", len(spirits))
	}
	for _, s := range spirits {
		if p, tg := effectstest.SocB14PT(t, g, s.InstanceID); p != 3 || tg != 3 {
			t.Errorf("Spirit token is %d/%d, want 3/3 after the flashback counter", p, tg)
		}
	}
	if n := effectstest.CountersOn(g, older, game.CounterPlusOne); n != 1 {
		t.Errorf("the older Spirit has %d counters, want 1 (each Spirit you control)", n)
	}
	if n := effectstest.CountersOn(g, other, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Spirit has %d counters, want 0", n)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back card returned to the graveyard instead of being exiled")
	}
}

// --- Applied Geometry -----------------------------------------------------

func TestSocB19AppliedGeometryCopiesACreatureAsASixCounterFractal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", socB19PlainCreatureTypes, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Applied Geometry", "Sorcery", socB19GeometryOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	copies := effectstest.SocB19Named(g, me.ID, "Grizzly Bears")
	if len(copies) != 1 {
		t.Fatalf("token copies %d, want 1", len(copies))
	}
	c := copies[0]
	if p, tg := effectstest.SocB14PT(t, g, c.InstanceID); p != 6 || tg != 6 {
		t.Errorf("the copy is %d/%d, want 6/6 (a 0/0 with six counters)", p, tg)
	}
	if !c.HasSubtype("Fractal") || !c.HasSubtype("Bear") || !c.IsCreature() {
		t.Errorf("the copy's type line is %q, want a Bear and a Fractal creature", c.TypeLine)
	}
	if n := effectstest.CountersOn(g, c.InstanceID, game.CounterPlusOne); n != 6 {
		t.Errorf("counters %d, want 6", n)
	}
	// The original is untouched.
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 2 || tg != 2 {
		t.Errorf("the original is %d/%d, want 2/2", p, tg)
	}
}

func TestSocB19AppliedGeometryMakesANoncreaturePermanentAFractalCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Applied Geometry", "Sorcery", socB19GeometryOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	copies := effectstest.SocB19Named(g, me.ID, "Mind Stone")
	if len(copies) != 1 {
		t.Fatalf("token copies %d, want 1", len(copies))
	}
	c := copies[0]
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Fractal") {
		t.Errorf("the copy's type line is %q, want an artifact creature that is a Fractal", c.TypeLine)
	}
	if p, tg := effectstest.SocB14PT(t, g, c.InstanceID); p != 6 || tg != 6 {
		t.Errorf("the copy is %d/%d, want 6/6", p, tg)
	}
}

func TestSocB19AppliedGeometryRefusesAnAuraAndAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.B12Permanent(g, me.ID, "Pacifism", "Enchantment — Aura")
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	for what, target := range map[string]uuid.UUID{"an Aura": aura, "an opponent's creature": theirs} {
		err := effectstest.CastCatalogSpellErr(t, g, "Applied Geometry", "Sorcery", socB19GeometryOracle, effectstest.CardRefs(target))
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("Applied Geometry at %s: err = %v, want ErrIllegalTarget", what, err)
		}
	}
}
