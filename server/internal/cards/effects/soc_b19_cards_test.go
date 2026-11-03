package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB19CastWithPool puts a catalog card in the active seat's hand, floats
// `color` x `n` in its pool in a main phase and casts it Strict.
func socB19CastWithPool(g *game.Game, me *game.Player, c game.Card, color string, n int, params game.CastSpellParams) error {
	c.Owner, c.Controller = me.ID, me.ID
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	me.Hand.PushTop(c)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, color, n)
	params.Strict = true
	return g.CastSpell(me.ID, c.InstanceID, params)
}

// --- Aberrant Manawurm ----------------------------------------------------

func TestSocB19ManawurmGetsPlusXForTheManaSpentOnAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worm := b12Push(g, me.ID, "Aberrant Manawurm", "Creature — Wurm", socB19ManawurmOracle, 2, 5)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{G}"}, "G", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, worm); p != 5 {
		t.Errorf("power %d after a three-mana sorcery, want 2+3", p)
	}
	if err := socB19CastWithPool(g, me, game.Card{Name: "One", TypeLine: "Instant", ManaCost: "{G}"}, "G", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, worm); p != 6 {
		t.Errorf("power %d after a second one-mana instant, want 6", p)
	}
	// Until end of turn only.
	endTurn(t, g)
	if p, _ := socB14PT(t, g, worm); p != 2 {
		t.Errorf("power %d on the next turn, want it back at 2", p)
	}
}

func TestSocB19ManawurmIgnoresACreatureSpellAndAFreeSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	worm := b12Push(g, me.ID, "Aberrant Manawurm", "Creature — Wurm", socB19ManawurmOracle, 2, 5)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: socB19PlainCreatureTypes, ManaCost: "{1}{G}", Power: 2, Toughness: 2}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, worm); p != 2 {
		t.Errorf("power %d after a creature spell, want 2", p)
	}
	// A spell cast without paying spent no mana: +0/+0.
	socB4Spell(t, g, "Free", "Instant", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, worm); p != 2 {
		t.Errorf("power %d after a free instant, want 2", p)
	}
}

// --- Abigale, Poet Laureate // Heroic Stanza ------------------------------

func socB19CastAbigale(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB19AbigaleOracle, []string{"flying"}, "Abigale, Poet Laureate", "Legendary Creature — Bird Bard", "{1}{W}{B}", 2, 3,
		"Heroic Stanza", "Sorcery", "{1}{W/B}")
	castFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func TestSocB19AbigalePreparesWhenYouCastACreatureSpellAndStanzaPutsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	abigale := socB19CastAbigale(t, g, me)
	if g.IsPreparedForEffect(abigale) {
		t.Fatal("prepared on entering")
	}
	// A noncreature spell does not prepare her.
	socB4Spell(t, g, "Bolt", "Instant", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if g.IsPreparedForEffect(abigale) {
		t.Fatal("prepared by an instant")
	}
	bear := b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	socB4Spell(t, g, "Elk", "Creature — Elk", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(abigale) {
		t.Fatal("not prepared after casting a creature spell")
	}
	cp, ok := prepareCopyOf(g, "Heroic Stanza")
	if !ok {
		t.Fatal("no Heroic Stanza copy in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast Heroic Stanza: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters on the target %d, want 1", n)
	}
	if g.IsPreparedForEffect(abigale) {
		t.Error("still prepared after casting the copy")
	}
}

// --- Abstract Paintmage ---------------------------------------------------

func TestSocB19PaintmageAddsURForInstantsAndSorceriesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Abstract Paintmage", "Creature — Djinn Sorcerer", socB19PaintmageOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pm := b12Push(g, me.ID, "Abstract Paintmage", "Creature — Djinn Sorcerer", socB19PaintmageOracle, 2, 2)
	endTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	if n := triggersOnStackFrom(g, pm); n != 0 {
		t.Errorf("%d Paintmage triggers on an opponent's main phase", n)
	}
}

// --- Additive Evolution ---------------------------------------------------

func TestSocB19AdditiveEvolutionMakesAThreeCounterFractalAndGrowsACreatureEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	castCatalogSpell(t, g, "Additive Evolution", "Enchantment", socB19EvolutionOracle, nil)
	passPriorityAroundTable(t, g)
	fractals := socB4Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("Fractals %d, want 1", len(fractals))
	}
	if p, tg := socB14PT(t, g, fractals[0].InstanceID); p != 3 || tg != 3 {
		t.Errorf("Fractal is %d/%d, want 3/3", p, tg)
	}
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("Bear is %d/%d after the combat trigger, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the target did not gain vigilance")
	}
}

func TestSocB19AdditiveEvolutionOnlyTriggersOnYourOwnCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Additive Evolution", "Enchantment", socB19EvolutionOracle, false)
	b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	endTurn(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	if n := triggersOnStackFrom(g, id); n != 0 {
		t.Errorf("%d triggers at an opponent's beginning of combat", n)
	}
}

// --- Adventurous Eater // Have a Bite -------------------------------------

func TestSocB19AdventurousEaterEntersPreparedAndHaveABiteCountersAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB19EaterOracle, nil, "Adventurous Eater", "Creature — Human Warlock", "{2}{B}", 3, 2,
		"Have a Bite", "Sorcery", "{B}")
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("did not enter prepared")
	}
	cp, ok := prepareCopyOf(g, "Have a Bite")
	if !ok {
		t.Fatal("no Have a Bite copy in exile")
	}
	bear := b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	life := me.Life
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast Have a Bite: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, tapped) })
	advanceToMain(t, g)
	card := game.Card{Name: "Ajani's Response", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: socB19ResponseOracle}

	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("cast at the full price with two mana against an untapped creature")
	}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature died")
	}
}

func TestSocB19AjanisResponseAtFullPriceDestroysAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	advanceToMain(t, g)
	card := game.Card{Name: "Ajani's Response", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: socB19ResponseOracle}
	if err := socB19CastWithPool(g, me, card, "W", 5, game.CastSpellParams{Targets: cardRefs(ox)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ox) {
		t.Error("the creature survived")
	}
}

// --- Ambitious Augmenter --------------------------------------------------

func socB19Augmenter(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Ambitious Augmenter", "Creature — Turtle Wizard", socB19AugmenterOracle, 1, 1)
}

func TestSocB19AugmenterIncrementsOnlyWhenTheManaSpentExceedsPowerOrToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aug := socB19Augmenter(g, me)
	advanceToMain(t, g)
	cast := func(cost string, n int) {
		t.Helper()
		if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: cost}, "G", n, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		passPriorityAroundTable(t, g)
	}
	cast("{G}", 1) // 1 is not greater than 1/1
	if n := countersOn(g, aug, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters %d after a one-mana spell, want 0", n)
	}
	cast("{1}{G}", 2) // 2 > 1: now 2/2
	if n := countersOn(g, aug, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a two-mana spell, want 1", n)
	}
	cast("{1}{G}", 2) // 2 is not greater than 2/2
	if n := countersOn(g, aug, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a second two-mana spell, want still 1", n)
	}
	cast("{2}{G}", 3) // 3 > 2: now 3/3
	if n := countersOn(g, aug, game.CounterPlusOne); n != 2 {
		t.Fatalf("counters %d after a three-mana spell, want 2", n)
	}
}

func TestSocB19AugmenterRechecksTheConditionAsTheTriggerResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aug := socB19Augmenter(g, me)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}{G}"}, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := triggersOnStackFrom(g, aug); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	// In response it grows to 2/2: two mana is no longer greater (CR 603.4).
	addCountersForTest(g, aug, game.CounterPlusOne, 1)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, aug, game.CounterPlusOne); n != 1 {
		t.Errorf("counters %d, want the one added in response and no more", n)
	}
}

func TestSocB19AugmenterDiesWithCountersAndTheFractalTakesThemAll(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aug := socB19Augmenter(g, me)
	addCountersForTest(g, aug, game.CounterPlusOne, 2)
	addCountersForTest(g, aug, "stun", 1)
	_ = opp
	castCatalogSpell(t, g, "Ajani's Response", "Instant", socB19ResponseOracle, cardRefs(aug))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aug) {
		t.Fatal("the Augmenter survived")
	}
	fractals := socB4Fractals(g, me.ID)
	if len(fractals) != 1 {
		t.Fatalf("Fractals %d, want 1", len(fractals))
	}
	id := fractals[0].InstanceID
	if n := countersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters on the Fractal %d, want 2", n)
	}
	if n := countersOn(g, id, "stun"); n != 1 {
		t.Errorf("stun counters on the Fractal %d, want 1", n)
	}
	if p, tg := socB14PT(t, g, id); p != 2 || tg != 2 {
		t.Errorf("Fractal is %d/%d, want 2/2", p, tg)
	}
}

func TestSocB19AugmenterDiesWithoutCountersMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aug := socB19Augmenter(g, me)
	castCatalogSpell(t, g, "Ajani's Response", "Instant", socB19ResponseOracle, cardRefs(aug))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aug) {
		t.Fatal("the Augmenter survived")
	}
	if n := len(socB4Fractals(g, me.ID)); n != 0 {
		t.Errorf("Fractals %d, want none for a creature that died with no counters", n)
	}
}

// --- Ancestral Anger ------------------------------------------------------

func TestSocB19AncestralAngerCountsCardsNamedAncestralAngerInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ancestral Anger", TypeLine: "Sorcery", OracleID: socB19AngerOracle, Owner: me.ID, Controller: me.ID})
	}
	// Another card in the graveyard does not count.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Other", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	hand := me.Hand.Size()
	card := game.Card{Name: "Ancestral Anger", TypeLine: "Sorcery", ManaCost: "{R}", OracleID: socB19AngerOracle}
	if err := socB19CastWithPool(g, me, card, "R", 1, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, bear); p != 5 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 5/2 (X = 1 + 2)", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the spell draws a card", me.Hand.Size(), hand+1)
	}
	// Until end of turn only.
	endTurn(t, g)
	if p, _ := socB14PT(t, g, bear); p != 2 {
		t.Errorf("power %d next turn, want 2", p)
	}
}

func TestSocB19AncestralAngerAloneIsPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	castCatalogSpell(t, g, "Ancestral Anger", "Sorcery", socB19AngerOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3 (X = 1)", p)
	}
}

// --- Antiquities on the Loose ---------------------------------------------

func TestSocB19AntiquitiesFromHandMakesTwoPlainSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	card := game.Card{Name: "Antiquities on the Loose", TypeLine: "Sorcery", ManaCost: "{1}{W}{W}", OracleID: socB19AntiquitiesOracle}
	if err := socB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	spirits := socB19Named(g, me.ID, "Spirit")
	if len(spirits) != 2 {
		t.Fatalf("Spirits %d, want 2", len(spirits))
	}
	for _, s := range spirits {
		if p, tg := socB14PT(t, g, s.InstanceID); p != 2 || tg != 2 {
			t.Errorf("Spirit is %d/%d, want 2/2 with no counter from a hand cast", p, tg)
		}
	}
}

func TestSocB19AntiquitiesFlashbackPutsACounterOnEachSpiritYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	older := b12Creature(g, me.ID, "Old Spirit", "Creature — Spirit", 1, 1)
	other := b12Creature(g, me.ID, "Bear", socB19PlainCreatureTypes, 2, 2)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Antiquities on the Loose", TypeLine: "Sorcery", ManaCost: "{1}{W}{W}",
		OracleID: socB19AntiquitiesOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "W", 6)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	passPriorityAroundTable(t, g)
	spirits := socB19Named(g, me.ID, "Spirit")
	if len(spirits) != 2 {
		t.Fatalf("token Spirits %d, want 2", len(spirits))
	}
	for _, s := range spirits {
		if p, tg := socB14PT(t, g, s.InstanceID); p != 3 || tg != 3 {
			t.Errorf("Spirit token is %d/%d, want 3/3 after the flashback counter", p, tg)
		}
	}
	if n := countersOn(g, older, game.CounterPlusOne); n != 1 {
		t.Errorf("the older Spirit has %d counters, want 1 (each Spirit you control)", n)
	}
	if n := countersOn(g, other, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Spirit has %d counters, want 0", n)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back card returned to the graveyard instead of being exiled")
	}
}

// socB19Named lists the battlefield tokens or cards of that name `owner` controls.
func socB19Named(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

// --- Applied Geometry -----------------------------------------------------

func TestSocB19AppliedGeometryCopiesACreatureAsASixCounterFractal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", socB19PlainCreatureTypes, 2, 2)
	castCatalogSpell(t, g, "Applied Geometry", "Sorcery", socB19GeometryOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	copies := socB19Named(g, me.ID, "Grizzly Bears")
	if len(copies) != 1 {
		t.Fatalf("token copies %d, want 1", len(copies))
	}
	c := copies[0]
	if p, tg := socB14PT(t, g, c.InstanceID); p != 6 || tg != 6 {
		t.Errorf("the copy is %d/%d, want 6/6 (a 0/0 with six counters)", p, tg)
	}
	if !c.HasSubtype("Fractal") || !c.HasSubtype("Bear") || !c.IsCreature() {
		t.Errorf("the copy's type line is %q, want a Bear and a Fractal creature", c.TypeLine)
	}
	if n := countersOn(g, c.InstanceID, game.CounterPlusOne); n != 6 {
		t.Errorf("counters %d, want 6", n)
	}
	// The original is untouched.
	if p, tg := socB14PT(t, g, bear); p != 2 || tg != 2 {
		t.Errorf("the original is %d/%d, want 2/2", p, tg)
	}
}

func TestSocB19AppliedGeometryMakesANoncreaturePermanentAFractalCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	castCatalogSpell(t, g, "Applied Geometry", "Sorcery", socB19GeometryOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	copies := socB19Named(g, me.ID, "Mind Stone")
	if len(copies) != 1 {
		t.Fatalf("token copies %d, want 1", len(copies))
	}
	c := copies[0]
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Fractal") {
		t.Errorf("the copy's type line is %q, want an artifact creature that is a Fractal", c.TypeLine)
	}
	if p, tg := socB14PT(t, g, c.InstanceID); p != 6 || tg != 6 {
		t.Errorf("the copy is %d/%d, want 6/6", p, tg)
	}
}

func TestSocB19AppliedGeometryRefusesAnAuraAndAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := b12Permanent(g, me.ID, "Pacifism", "Enchantment — Aura")
	theirs := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	for what, target := range map[string]uuid.UUID{"an Aura": aura, "an opponent's creature": theirs} {
		err := castCatalogSpellErr(t, g, "Applied Geometry", "Sorcery", socB19GeometryOracle, cardRefs(target))
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("Applied Geometry at %s: err = %v, want ErrIllegalTarget", what, err)
		}
	}
}
