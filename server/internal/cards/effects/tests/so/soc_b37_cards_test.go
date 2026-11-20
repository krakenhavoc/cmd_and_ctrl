package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b37_cards_test.go — Secrets of Strixhaven (sos), batch 37 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB37SunderingOracle = "b095526e-94a4-416b-83de-d6271804ccf3"
	socB37SuspendOracle   = "883267f0-be7e-43c3-b7f8-82981f6d8c23"
	socB37TackleOracle    = "2e80d632-cf84-4562-8c93-5850cceb5bec"
	socB37TamOracle       = "702e871d-90d8-4468-8f69-5ae42af2c9d3"
	socB37TeachersOracle  = "12397dbf-99a9-40c3-aa7c-76d09baad008"
	socB37ConcocterOracle = "0ecd4493-dfb4-4d56-a90e-08b74b474f39"
	socB37TesterOracle    = "a30a2b96-4a8c-4d18-b276-5005b8aff7ed"
	socB37TabulatorOracle = "ebb40a50-cfbc-4ec1-9c54-26a1bc2e7a6e"
	socB37DawningOracle   = "a3e10b9b-9349-4b44-a46c-c825293dbd05"
	socB37KelpieOracle    = "a2db1035-008e-4de3-b5e5-2a0de83082b2"
	socB37PlusOne         = game.CounterPlusOne
)

func TestSocB37EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB37SunderingOracle:  "Sundering Archaic",
		socB37SuspendOracle:    "Suspend Aggression",
		socB37TackleOracle:     "Tackle Artist",
		socB37TamOracle:        "Tam, Observant Sequencer",
		socB37TamOracle + "#1": "Deep Sight",
		socB37TeachersOracle:   "Teacher's Pest",
		socB37ConcocterOracle:  "Tenured Concocter",
		socB37TesterOracle:     "Tester of the Tangential",
		socB37TabulatorOracle:  "Textbook Tabulator",
		socB37DawningOracle:    "The Dawning Archaic",
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

// socB37Cast casts a small spell with exactly the mana it costs.
func socB37Cast(t *testing.T, g *game.Game, me *game.Player, typeLine, cost string, n int) {
	t.Helper()
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: typeLine, ManaCost: cost}, "R", n, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Tackle Artist -----------------------------------------------------------

func TestSocB37TackleArtistGrowsByOneOrTwoAndIgnoresCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	artist := effectstest.B12Push(g, me.ID, "Tackle Artist", "Creature — Orc Sorcerer", socB37TackleOracle, 4, 3)
	effectstest.AdvanceToMain(t, g)
	socB37Cast(t, g, me, "Instant", "{3}{R}", 4)
	if n := effectstest.CountersOn(g, artist, socB37PlusOne); n != 1 {
		t.Fatalf("counters %d after a four-mana instant, want 1", n)
	}
	socB37Cast(t, g, me, "Sorcery", "{4}{R}", 5)
	if n := effectstest.CountersOn(g, artist, socB37PlusOne); n != 3 {
		t.Fatalf("counters %d after a five-mana sorcery, want 1+2", n)
	}
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{R}", Power: 2, Toughness: 2}, "R", 5, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, artist, socB37PlusOne); n != 3 {
		t.Errorf("counters %d after a creature spell, want 3", n)
	}
}

func TestSocB37TackleArtistHasTrample(t *testing.T) {
	spec, ok := effects.Lookup(socB37TackleOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "trample" {
		t.Fatalf("keywords %v, want trample", spec.PrintedKeywords)
	}
}

// --- Textbook Tabulator ------------------------------------------------------

func TestSocB37TabulatorIncrementsOnlyPastItsPowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tab := effectstest.B12Push(g, me.ID, "Textbook Tabulator", "Creature — Frog Wizard", socB37TabulatorOracle, 0, 3)
	effectstest.AdvanceToMain(t, g)
	socB37Cast(t, g, me, "Instant", "{R}", 1)
	if n := effectstest.CountersOn(g, tab, socB37PlusOne); n != 1 {
		t.Fatalf("counters %d after one mana against power 0, want 1", n)
	}
	// Now 1/4: one mana is not greater than the power of 1.
	socB37Cast(t, g, me, "Instant", "{R}", 1)
	if n := effectstest.CountersOn(g, tab, socB37PlusOne); n != 1 {
		t.Fatalf("counters %d after one mana against power 1, want 1", n)
	}
	socB37Cast(t, g, me, "Instant", "{1}{R}", 2)
	if n := effectstest.CountersOn(g, tab, socB37PlusOne); n != 2 {
		t.Fatalf("counters %d after two mana against power 1, want 2", n)
	}
}

func TestSocB37TabulatorSurveilsTwoWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bin Me", "Keep Me")
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Textbook Tabulator", TypeLine: "Creature — Frog Wizard", OracleID: socB37TabulatorOracle,
		Power: 0, Toughness: 3})
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not queue a surveil")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil looks at %d cards, want 2", len(c.ScryCards))
	}
}

// --- Tester of the Tangential --------------------------------------------

func socB37Tester(t *testing.T, counters int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tester := effectstest.B12Push(g, me.ID, "Tester of the Tangential", "Creature — Djinn Wizard", socB37TesterOracle, 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if counters > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(tester, socB37PlusOne, counters) })
	}
	return g, me, tester, bear
}

func TestSocB37TesterIncrementPutsACounter(t *testing.T) {
	g, me, tester, _ := socB37Tester(t, 0)
	effectstest.AdvanceToMain(t, g)
	socB37Cast(t, g, me, "Instant", "{1}{R}", 2)
	if n := effectstest.CountersOn(g, tester, socB37PlusOne); n != 1 {
		t.Fatalf("counters %d, want 1", n)
	}
}

func TestSocB37TesterPaysXAtBeginningOfCombatAndMovesThatManyCounters(t *testing.T) {
	g, me, tester, bear := socB37Tester(t, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPool(me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	// The target is chosen only now, as the reflexive trigger goes on the
	// stack, and the Tester itself is not offered.
	ch := effectstest.PickTargetPrompt(t, g)
	for _, id := range ch.PickTargetCards {
		if id == tester {
			t.Error("the Tester is offered as 'another target creature'")
		}
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, tester, socB37PlusOne); n != 1 {
		t.Errorf("Tester has %d counters, want 3-2", n)
	}
	if n := effectstest.CountersOn(g, bear, socB37PlusOne); n != 2 {
		t.Errorf("the Bear has %d counters, want 2", n)
	}
}

func TestSocB37TesterPayingNothingMovesNothingAndAsksNoTarget(t *testing.T) {
	g, me, tester, bear := socB37Tester(t, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPool(me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("a target was asked for after paying nothing")
		}
	}
	if n := effectstest.CountersOn(g, tester, socB37PlusOne); n != 3 {
		t.Errorf("Tester has %d counters, want 3", n)
	}
	if n := effectstest.CountersOn(g, bear, socB37PlusOne); n != 0 {
		t.Errorf("the Bear has %d counters, want 0", n)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("pool %d, want the 2 mana untouched", got)
	}
}

func TestSocB37TesterCannotMoveMoreThanItHas(t *testing.T) {
	g, me, tester, bear := socB37Tester(t, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPool(me, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, tester, socB37PlusOne); n != 0 {
		t.Errorf("Tester has %d counters, want 0", n)
	}
	if n := effectstest.CountersOn(g, bear, socB37PlusOne); n != 1 {
		t.Errorf("the Bear has %d counters, want only the 1 there was", n)
	}
}

// --- Teacher's Pest ----------------------------------------------------------

func TestSocB37TeachersPestGainsALifeWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pest := effectstest.B12Push(g, me.ID, "Teacher's Pest", "Creature — Skeleton Pest", socB37TeachersOracle, 1, 1)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, pest)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestSocB37TeachersPestReturnsFromTheGraveyardTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Teacher's Pest", "Creature — Skeleton Pest", socB37TeachersOracle, 1, 1,
		"{B}{G}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the Pest did not return to the battlefield")
	}
	c, _ := g.LookupCardForEffect(id)
	if !c.Tapped {
		t.Error("the Pest must enter tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("it entered via a replacement, not a tap: %d EventTapCard", n)
	}
}

func TestSocB37TeachersPestCantReturnWithoutTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, "Teacher's Pest", "Creature — Skeleton Pest", socB37TeachersOracle, 1, 1)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only {B}")
	}
	if g.Battlefield.Contains(id) {
		t.Error("the Pest returned without paying {B}{G}")
	}
}

// --- Tenured Concocter -----------------------------------------------------

func socB37TargetWith(t *testing.T, g *game.Game, who *game.Player, target uuid.UUID) {
	t.Helper()
	spell := game.NewCard("Test Removal", who.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{R}"
	who.Hand.PushTop(spell)
	if err := g.CastSpell(who.ID, spell.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("cast the targeting spell: %v", err)
	}
}

func TestSocB37ConcocterMayDrawWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conc := effectstest.B12Push(g, me.ID, "Tenured Concocter", "Creature — Troll Druid", socB37ConcocterOracle, 4, 5)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	socB37TargetWith(t, g, opp, conc)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("being targeted by an opponent offers a draw")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one drawn", hand, me.Hand.Size())
	}
}

func TestSocB37ConcocterMayDeclineTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conc := effectstest.B12Push(g, me.ID, "Tenured Concocter", "Creature — Troll Druid", socB37ConcocterOracle, 4, 5)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	socB37TargetWith(t, g, opp, conc)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want nothing drawn", hand, me.Hand.Size())
	}
}

func TestSocB37ConcocterIgnoresItsOwnControllersTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	conc := effectstest.B12Push(g, me.ID, "Tenured Concocter", "Creature — Troll Druid", socB37ConcocterOracle, 4, 5)
	effectstest.AdvanceToMain(t, g)
	socB37TargetWith(t, g, me, conc)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("targeting it yourself offers a draw")
	}
}

func TestSocB37ConcocterInfusionGivesPlusTwoPowerOnlyAfterYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conc := effectstest.B12Push(g, me.ID, "Tenured Concocter", "Creature — Troll Druid", socB37ConcocterOracle, 4, 5)
	effectstest.AdvanceToMain(t, g)
	if p, _ := effectstest.SocB14PT(t, g, conc); p != 4 {
		t.Fatalf("power %d before any life gain, want 4", p)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	if p, _ := effectstest.SocB14PT(t, g, conc); p != 4 {
		t.Errorf("power %d after an opponent gained life, want 4", p)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	if p, tough := effectstest.SocB14PT(t, g, conc); p != 6 || tough != 5 {
		t.Errorf("%d/%d after you gained life, want 6/5", p, tough)
	}
}

// --- Tam, Observant Sequencer // Deep Sight -------------------------------

func socB37TamCard(owner uuid.UUID) game.Card {
	return sos1Prep(owner, socB37TamOracle, nil, "Tam, Observant Sequencer", "Legendary Creature — Gorgon Wizard", "{2}{G}{U}", 4, 3,
		"Deep Sight", "Sorcery", "{G}{U}")
}

func TestSocB37TamBecomesPreparedOnLandfallAndDeepSightDrawsAndGains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := socB37TamCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Tam enters unprepared")
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("a land entering did not prepare Tam")
	}
	cp, ok := effectstest.PrepareCopyOf(g, "Deep Sight")
	if !ok {
		t.Fatal("no Deep Sight copy in exile")
	}
	hand, life := me.Hand.Size(), me.Life
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast Deep Sight: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Errorf("hand %d -> %d, life %d -> %d, want a card and 1 life", hand, me.Hand.Size(), life, me.Life)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB37TamIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := socB37TamCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	land := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(land)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, land.InstanceID); err != nil {
		t.Fatalf("put the land onto the battlefield: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBoard(g, land.InstanceID) {
		t.Fatal("the opponent's land did not enter")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("an opponent's land prepared Tam")
	}
}

// --- Sundering Archaic ---------------------------------------------------

// socB37CastArchaic casts Sundering Archaic for six mana of the given
// colours and returns its id.
func socB37CastArchaic(t *testing.T, g *game.Game, me *game.Player, colors string) {
	t.Helper()
	c := game.Card{Name: "Sundering Archaic", TypeLine: "Creature — Avatar", ManaCost: "{6}", OracleID: socB37SunderingOracle,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, InstanceID: uuid.New()}
	me.Hand.PushTop(c)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	for _, r := range colors {
		effectstest.FillPoolColored(me, string(r), 1)
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Sundering Archaic: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB37SunderingArchaicOffersOnlyPermanentsWithinTheColorsSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.PushPermanentForTest(g, opp.ID, "Cheap Relic", "", "Artifact")
	setManaValue(g, cheap, "{3}")
	dear := effectstest.PushPermanentForTest(g, opp.ID, "Dear Relic", "", "Artifact")
	setManaValue(g, dear, "{4}")
	land := effectstest.PushPermanentForTest(g, opp.ID, "Their Forest", "", "Basic Land — Forest")
	mine := effectstest.PushPermanentForTest(g, me.ID, "My Relic", "", "Artifact")
	setManaValue(g, mine, "{1}")
	effectstest.AdvanceToMain(t, g)
	socB37CastArchaic(t, g, me, "WWUUBB")
	var offered map[uuid.UUID]bool
	if ch := effectstest.SocB37OpenPickTarget(g); ch != nil {
		offered = map[uuid.UUID]bool{}
		for _, id := range ch.PickTargetCards {
			offered[id] = true
		}
	} else {
		t.Fatal("no target prompt: expected the three-colour clause to offer the cheap relic and more")
	}
	for id, want := range map[uuid.UUID]bool{cheap: true, dear: false, land: false, mine: false} {
		if offered[id] != want {
			t.Errorf("%s offered = %v, want %v", id, offered[id], want)
		}
	}
	effectstest.AnswerPickTarget(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if got := socB23ZoneOf(g, cheap); got != game.ZoneExile {
		t.Errorf("the cheap relic is in %v, want exile", got)
	}
	if !effectstest.OnBoard(g, dear) {
		t.Error("the expensive relic was exiled")
	}
}

func TestSocB37SunderingArchaicOneColourReachesOnlyManaValueOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := effectstest.PushPermanentForTest(g, opp.ID, "One Relic", "", "Artifact")
	setManaValue(g, one, "{1}")
	two := effectstest.PushPermanentForTest(g, opp.ID, "Two Relic", "", "Artifact")
	setManaValue(g, two, "{2}")
	effectstest.AdvanceToMain(t, g)
	socB37CastArchaic(t, g, me, "UUUUUU")
	// Only one target is legal, so the trigger may pick it for the player.
	if ch := effectstest.SocB37OpenPickTarget(g); ch != nil {
		effectstest.AnswerPickTarget(t, g, one)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := socB23ZoneOf(g, one); got != game.ZoneExile {
		t.Errorf("the mana value 1 relic is in %v, want exile", got)
	}
	if !effectstest.OnBoard(g, two) {
		t.Error("the mana value 2 relic was exiled with a single colour spent")
	}
}

func TestSocB37SunderingArchaicWithNothingLegalDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dear := effectstest.PushPermanentForTest(g, opp.ID, "Dear Relic", "", "Artifact")
	setManaValue(g, dear, "{5}")
	effectstest.AdvanceToMain(t, g)
	socB37CastArchaic(t, g, me, "UUUUUU")
	if !effectstest.OnBoard(g, dear) {
		t.Error("an out-of-range permanent was exiled")
	}
	if effectstest.SocB37OpenPickTarget(g) != nil {
		t.Error("a target prompt opened with nothing legal")
	}
}

func TestSocB37SunderingArchaicPutsAGraveyardCardOnTheBottomOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.GyCard(opp, "Their Card", "Sorcery", "{R}", "")
	archaic := effectstest.B12Push(g, me.ID, "Sundering Archaic", "Creature — Avatar", socB37SunderingOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, archaic, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) {
		t.Fatal("the card is still in the graveyard")
	}
	n := len(opp.Library.Cards)
	if n == 0 || opp.Library.Cards[n-1].InstanceID != card && opp.Library.Cards[0].InstanceID != card {
		t.Error("the card is not at either end of its owner's library")
	}
}

// --- Suspend Aggression --------------------------------------------------

func TestSocB37SuspendAggressionExilesBothAndEachOwnerMayPlayTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushPermanentForTest(g, opp.ID, "Their Relic", "", "Artifact")
	top := effectstest.SeedLibraryTop(me, "Top Land", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Suspend Aggression", "Instant", socB37SuspendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := socB23ZoneOf(g, theirs); got != game.ZoneExile {
		t.Fatalf("the permanent is in %v, want exile", got)
	}
	if got := socB23ZoneOf(g, top); got != game.ZoneExile {
		t.Fatalf("the top card is in %v, want exile", got)
	}
	permT := effectstest.ExiledPermission(g, theirs)
	if permT.Player != opp.ID || permT.CastOnly {
		t.Errorf("the permanent's permission %+v, want a play permission for its owner", permT)
	}
	if !effectstest.PermissionLive(g, permT, opp.ID) {
		t.Error("the opponent's permission is not live")
	}
	permTop := effectstest.ExiledPermission(g, top)
	if permTop.Player != me.ID || permTop.CastOnly {
		t.Errorf("the top card's permission %+v, want a play permission for the caster", permTop)
	}
	// The opponent's next turn is seat 1's: their permission lasts through it.
	effectstest.EndTurn(t, g)
	if g.CastPermissionOnCardByIDForEffect(theirs) == nil {
		t.Error("the opponent's permission lapsed before their turn")
	}
	effectstest.EndTurn(t, g)
	if g.CastPermissionOnCardByIDForEffect(theirs) != nil {
		t.Error("the opponent's permission outlasted the end of their next turn")
	}
	if g.CastPermissionOnCardByIDForEffect(top) == nil {
		t.Error("the caster's permission lapsed before the end of their next turn")
	}
}

func TestSocB37SuspendAggressionRefusesALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.PushPermanentForTest(g, opp.ID, "Their Forest", "", "Basic Land — Forest")
	err := effectstest.CastCatalogSpellErr(t, g, "Suspend Aggression", "Instant", socB37SuspendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	if err == nil {
		t.Error("a land was accepted as the target of a nonland-permanent spell")
	}
}

func TestSocB37SuspendAggressionExilesATokenForGoodAndStillTakesTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	token := effectstest.PushToken(g, opp.ID, effects.FoodToken())
	top := effectstest.SeedLibraryTop(me, "Top Card", "Sorcery", "{R}")
	effectstest.CastCatalogSpell(t, g, "Suspend Aggression", "Instant", socB37SuspendOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: token}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBoard(g, token) {
		t.Error("the token is still on the battlefield")
	}
	if got := socB23ZoneOf(g, top); got != game.ZoneExile {
		t.Errorf("the top card is in %v, want exile", got)
	}
}

// --- The Dawning Archaic -----------------------------------------------

func TestSocB37DawningArchaicCostsOneLessPerInstantAndSorceryInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.GyCard(me, "Old Instant", "Instant", "{R}", "")
	effectstest.GyCard(me, "Old Sorcery", "Sorcery", "{R}", "")
	effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	c := game.Card{Name: "The Dawning Archaic", TypeLine: "Legendary Creature — Avatar", ManaCost: "{10}", OracleID: socB37DawningOracle,
		Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID, InstanceID: uuid.New()}
	me.Hand.PushTop(c)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPool(me, 7)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for 7 with only two instant/sorcery cards in the graveyard")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPool(me, 8)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {10} minus 2: %v", err)
	}
}

func TestSocB37DawningArchaicAttackCastsAGraveyardSpellFreeAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archaic := effectstest.B12Push(g, me.ID, "The Dawning Archaic", "Legendary Creature — Avatar", socB37DawningOracle, 7, 7)
	card := effectstest.GyCard(me, "Draw Two", "Sorcery", "{2}{U}", "")
	effectstest.DeclareAttack(t, g, opp.ID, archaic)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, card)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, card, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast the graveyard card: %v", err)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		effectstest.DeclineResolutionCast(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := socB23ZoneOf(g, card); got != game.ZoneExile {
		t.Errorf("the spell ended in %v, want exile instead of the graveyard", got)
	}
}

func TestSocB37DawningArchaicAttackOffersNoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archaic := effectstest.B12Push(g, me.ID, "The Dawning Archaic", "Legendary Creature — Avatar", socB37DawningOracle, 7, 7)
	effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.DeclareAttack(t, g, opp.ID, archaic)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB37OpenPickTarget(g) != nil {
		t.Error("a target prompt opened with only a creature card in the graveyard")
	}
}

// setManaValue gives a test permanent a printed mana cost.
func setManaValue(g *game.Game, id uuid.UUID, cost string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ManaCost = cost
			}
		}
	})
}

// Grazing Kelpie shares the bottom-of-library body.
func TestSocB37KelpieStillSharesTheBottomOfLibraryBody(t *testing.T) {
	if _, ok := effects.Lookup(socB37KelpieOracle); !ok {
		t.Fatal("Grazing Kelpie is not registered")
	}
}
