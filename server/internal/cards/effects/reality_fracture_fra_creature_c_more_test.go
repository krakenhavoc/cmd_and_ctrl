package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_c_more_test.go — the second half of the
// fra-creature-c card tests (Hungering Puppetbeast through Mabel).

// --- Hungering Puppetbeast ----------------------------------------

func TestHungeringPuppetbeastMakesAHeartwood(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Hungering Puppetbeast", "Artifact Creature — Beast Construct", rfcPuppetbeastOracle, 5, 5)
	if n := len(battlefieldIDsNamed(g, "Heartwood")); n != 1 {
		t.Fatalf("Heartwood tokens = %d, want 1", n)
	}
}

func TestHungeringPuppetbeastSacrificesAnArtifactForACounterAndAKeyword(t *testing.T) {
	for i, kw := range []string{"trample", "hexproof", "haste"} {
		g := newCatalogGame(t)
		me, _ := rfcSeats(g)
		toMain(t, g)
		beast := pushDiesCreatureForTest(g, me.ID, "Hungering Puppetbeast", rfcPuppetbeastOracle, "Artifact Creature — Beast Construct", 5, 5)
		rock := pushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
		if err := g.ActivateCatalogAbility(me.ID, beast, 0, game.ActivateAbilityParams{Modes: []int{i}, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
			t.Fatalf("%s: activate: %v", kw, err)
		}
		passPriorityAroundTable(t, g)
		if onBattlefieldNow(g, rock) {
			t.Errorf("%s: the sacrificed artifact is still there", kw)
		}
		if got := rfcCountersOf(t, g, beast, game.CounterPlusOne); got != 1 {
			t.Errorf("%s: counters = %d, want 1", kw, got)
		}
		if !hasAbility(effectiveAbilities(t, g, beast), kw) {
			t.Errorf("%s: keyword missing from %v", kw, effectiveAbilities(t, g, beast))
		}
	}
}

func TestHungeringPuppetbeastCannotSacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	toMain(t, g)
	beast := pushDiesCreatureForTest(g, me.ID, "Hungering Puppetbeast", rfcPuppetbeastOracle, "Artifact Creature — Beast Construct", 5, 5)
	err := g.ActivateCatalogAbility(me.ID, beast, 0, game.ActivateAbilityParams{Modes: []int{0}, SacrificeIDs: []uuid.UUID{beast}})
	if err == nil {
		t.Fatal("sacrificing itself must be refused (\"another artifact\")")
	}
}

// --- Ingris Stingerquill ------------------------------------------

func TestIngrisPingsEachOpponentPerAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Ingris Stingerquill", rfcIngrisOracle, rfcLegendaryCreature, 1, 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	before := map[uuid.UUID]int{}
	for _, p := range opps {
		before[p.ID] = p.Life
	}
	declareAttack(t, g, opps[0].ID, a, b)
	passPriorityAroundTable(t, g)
	for _, p := range opps {
		if got := before[p.ID] - p.Life; got != 2 {
			t.Errorf("%s lost %d from the two triggers, want 2", p.Name, got)
		}
	}
}

func TestIngrisActivationMakesACadetAndGivesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	toMain(t, g)
	ingris := pushDiesCreatureForTest(g, me.ID, "Ingris Stingerquill", rfcIngrisOracle, rfcLegendaryCreature, 1, 4)
	if err := g.ActivateCatalogAbility(me.ID, ingris, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	cadets := battlefieldIDsNamed(g, "Cadet")
	if len(cadets) != 1 {
		t.Fatalf("Cadets = %d, want 1", len(cadets))
	}
	if !hasAbility(effectiveAbilities(t, g, cadets[0]), "haste") {
		t.Error("the new Cadet should have haste")
	}
	if !hasAbility(effectiveAbilities(t, g, ingris), "haste") {
		t.Error("Ingris should have haste too")
	}
}

// --- Jhoira, Weatherlight Corsair ---------------------------------

func TestJhoiraTakesTheFirstHistoricPermanentAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	opp := opps[0]
	unreached, legend, bear, forest := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	// Index 0 is the bottom: the Forest is on top, then a Bear, then the
	// legend, then a card that is never reached.
	opp.Library.Cards = []game.Card{
		{InstanceID: unreached, Name: "Deep Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID},
		{InstanceID: legend, Name: "Old Legend", TypeLine: "Legendary Creature — Test", ManaCost: "{2}{R}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID},
		{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID},
		{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID},
	}
	lifeBefore := me.Life
	castHoldCreature(t, g, "Jhoira, Weatherlight Corsair", rfcLegendaryCreature, rfcJhoiraOracle, 4, 5)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, legend)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the legend should be on my side of the battlefield: %+v %v", c, ok)
	}
	if got := lifeBefore - me.Life; got != 3 {
		t.Errorf("life lost = %d, want 3 (the legend's mana value)", got)
	}
	for _, id := range []uuid.UUID{forest, bear} {
		if b12ZoneOf(g, id) != game.ZoneLibrary {
			t.Errorf("%s should be back in the library", id)
		}
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1]; top.InstanceID != unreached {
		t.Errorf("top of library = %s, want the unreached card (the revealed pair went to the bottom)", top.Name)
	}
}

func TestJhoiraWithNoHistoricCardBottomsEverythingAndLosesNoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	opp := opps[0]
	a, b := uuid.New(), uuid.New()
	opp.Library.Cards = []game.Card{
		{InstanceID: a, Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID},
		{InstanceID: b, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID},
	}
	lifeBefore := me.Life
	castHoldCreature(t, g, "Jhoira, Weatherlight Corsair", rfcLegendaryCreature, rfcJhoiraOracle, 4, 5)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if me.Life != lifeBefore {
		t.Errorf("life changed by %d with nothing to take", me.Life-lifeBefore)
	}
	if len(opp.Library.Cards) != 2 {
		t.Errorf("library has %d cards, want both back", len(opp.Library.Cards))
	}
}

// --- Jiang Yanggu, Alone ------------------------------------------

func TestJiangAloneLootsAndCountsDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	jiang := pushDiesCreatureForTest(g, me.ID, "Jiang Yanggu, Alone", rfcJiangAloneOracle, rfcLegendaryCreature, 4, 4)
	me.Hand.Cards = nil
	handCard(me, "Keep", "Sorcery")
	pitch := handCard(me, "Pitch", "Sorcery")
	declareAttack(t, g, opps[0].ID, jiang)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, jiang, game.CounterPlusOne); got != 1 {
		t.Errorf("counters = %d, want 1 (one card discarded this turn)", got)
	}
	if b12ZoneOf(g, pitch) != game.ZoneGraveyard {
		t.Error("the pitched card should be in the graveyard")
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want the kept card plus the drawn one", me.Hand.Size())
	}
}

func TestJiangAloneDoesNotTriggerWithTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	jiang := pushDiesCreatureForTest(g, me.ID, "Jiang Yanggu, Alone", rfcJiangAloneOracle, rfcLegendaryCreature, 4, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, opps[0].ID, jiang, bear)
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, me.ID); c != nil {
		t.Fatal("two attackers is not attacking alone")
	}
	if got := rfcCountersOf(t, g, jiang, game.CounterPlusOne); got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
}

// --- Jiang Yanggu, Never Alone ------------------------------------

func TestJiangNeverAloneMakesMowuAndUntapsTokensAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	castHoldCreature(t, g, "Jiang Yanggu, Never Alone", rfcLegendaryCreature, rfcJiangNeverOracle, 2, 2)
	mowus := battlefieldIDsNamed(g, "Mowu")
	if len(mowus) != 1 {
		t.Fatalf("Mowu tokens = %d, want 1", len(mowus))
	}
	if c, _ := battlefieldCard(g, mowus[0]); !c.HasSupertype("Legendary") || c.Power != 3 {
		t.Errorf("Mowu should be a legendary 3/3: %+v", c)
	}
	nonToken := pushCreatureToBattlefieldForTest(g, me.ID, "Plain Bear")
	g.WithWriteLock(func() {
		_ = g.TapTargetForEffect(mowus[0])
		_ = g.TapTargetForEffect(nonToken)
	})
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, mowus[0]); c.Tapped {
		t.Error("tokens untap at the beginning of your end step")
	}
	if c, _ := battlefieldCard(g, nonToken); !c.Tapped {
		t.Error("nontoken creatures stay tapped")
	}
}

// --- Karn, Argent Defender ----------------------------------------

func TestKarnArgentDefenderStopsCreatureEntryTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Karn, Argent Defender", rfcKarnDefenderOracle, "Legendary Artifact Creature — Golem", 1, 3)
	castHoldCreature(t, g, "Heartstring Puller", "Creature — Elf Sorcerer", rfcHeartstringOracle, 3, 1)
	if n := len(battlefieldIDsNamed(g, "Cadet")); n != 0 {
		t.Errorf("a creature's enters trigger should be suppressed, got %d Cadets", n)
	}
}

func TestKarnArgentDefenderStopsArtifactEntryTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Karn, Argent Defender", rfcKarnDefenderOracle, "Legendary Artifact Creature — Golem", 1, 3)
	// Hungering Puppetbeast is an artifact creature; use a non-creature
	// artifact source instead: Karn, Gilded Guardian's draw would fire.
	before := me.Hand.Size()
	castHoldCreature(t, g, "Karn, Gilded Guardian", "Legendary Artifact Creature — Golem", rfcKarnGuardianOracle, 5, 5)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("cards drawn = %d, want 0: the enters trigger is suppressed", got)
	}
}

func TestKarnArgentDefenderLeavesLandEntriesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Karn, Argent Defender", rfcKarnDefenderOracle, "Legendary Artifact Creature — Golem", 1, 3)
	koth := pushDiesCreatureForTest(g, me.ID, "Koth of the Homestead", rfcKothHomesteadOracle, rfcLegendaryCreature, 2, 3)
	before := me.Life
	rfcPlayLand(t, g, "Plains", "Basic Land — Plains")
	answerPickTarget(t, g, koth)
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("landfall life = %d, want %d: a land is neither an artifact nor a creature", me.Life, before+1)
	}
}

// --- Karn, Gilded Guardian ----------------------------------------

func TestKarnGildedGuardianDrawsPerColorAmongOtherArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	for _, colors := range [][]string{{"R"}, {"R", "G"}, nil, {"U"}} {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Gizmo", TypeLine: "Artifact", Colors: colors,
			Owner: me.ID, Controller: me.ID,
		})
	}
	before := me.Hand.Size()
	castHoldCreature(t, g, "Karn, Gilded Guardian", "Legendary Artifact Creature — Golem", rfcKarnGuardianOracle, 5, 5)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("cards drawn = %d, want 3 (R, G, U — Karn's own colours don't count)", got)
	}
}

// --- Kiora of Fire and Ashes --------------------------------------

func TestKioraMakesADragonOnEntryAndOnActivation(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	kiora := castHoldCreature(t, g, "Kiora of Fire and Ashes", rfcLegendaryCreature, rfcKioraOracle, 2, 2)
	dragons := battlefieldIDsNamed(g, "Dragon")
	if len(dragons) != 1 {
		t.Fatalf("Dragons after entry = %d, want 1", len(dragons))
	}
	if effectivePower(t, g, dragons[0]) != 5 || !hasAbility(effectiveAbilities(t, g, dragons[0]), "flying") {
		t.Error("the Dragon is a 5/5 flyer")
	}
	if err := g.ActivateCatalogAbility(me.ID, kiora, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Dragon")); n != 2 {
		t.Errorf("Dragons after {8} = %d, want 2", n)
	}
}

// --- Koth of the Homestead ----------------------------------------

func TestKothOfTheHomesteadGainsLifeAndPumpsOnPlains(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	koth := pushDiesCreatureForTest(g, me.ID, "Koth of the Homestead", rfcKothHomesteadOracle, rfcLegendaryCreature, 2, 3)
	before := me.Life
	rfcPlayLand(t, g, "Plains", "Basic Land — Plains")
	answerPickTarget(t, g, koth)
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life = %d, want %d", me.Life, before+1)
	}
	if got := rfcCountersOf(t, g, koth, game.CounterPlusOne); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

func TestKothOfTheHomesteadMountainOnlyGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	koth := pushDiesCreatureForTest(g, me.ID, "Koth of the Homestead", rfcKothHomesteadOracle, rfcLegendaryCreature, 2, 3)
	before := me.Life
	rfcPlayLand(t, g, "Mountain", "Basic Land — Mountain")
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life = %d, want %d", me.Life, before+1)
	}
	if got := rfcCountersOf(t, g, koth, game.CounterPlusOne); got != 0 {
		t.Errorf("a Mountain must not add a counter, got %d", got)
	}
}

// --- Koth, the Geomancer ------------------------------------------

func TestKothTheGeomancerLandfallPingsAndMountainAddsRed(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Koth, the Geomancer", rfcKothGeomancerOracle, rfcLegendaryCreature, 3, 2)
	before := opps[0].Life
	rfcPlayLand(t, g, "Forest", "Basic Land — Forest")
	passPriorityAroundTable(t, g)
	if got := before - opps[0].Life; got != 1 {
		t.Errorf("opponent lost %d, want 1", got)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("a Forest adds no mana, pool = %v", poolColors(me))
	}
	rfcPlayLand(t, g, "Mountain", "Basic Land — Mountain")
	passPriorityAroundTable(t, g)
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("a Mountain adds {R}, pool = %v", got)
	}
}

// --- Kwia Vigorbloom ----------------------------------------------

func TestKwiaMakesOneLotusPerTurnAndTheLotusMakesThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Kwia Vigorbloom", rfcKwiaOracle, rfcLegendaryCreature, 6, 6)
	rfcGainLife(g, me.ID, 2)
	passPriorityAroundTable(t, g)
	rfcGainLife(g, me.ID, 3)
	passPriorityAroundTable(t, g)
	lotuses := battlefieldIDsNamed(g, "Lotus")
	if len(lotuses) != 1 {
		t.Fatalf("Lotus tokens = %d, want 1 (triggers only once each turn)", len(lotuses))
	}
	if err := g.ActivateManaAbility(me.ID, lotuses[0], 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Lotus mana ability: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 || pick.ManaAmounts["W"] != 3 {
		t.Fatalf("pick = %+v, want one five-colour pick of three", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); len(got) != 3 {
		t.Errorf("pool = %v, want three mana", got)
	}
	if onBattlefieldNow(g, lotuses[0]) {
		t.Error("the Lotus is sacrificed for its mana")
	}
}

// --- Loot, the Nexus ----------------------------------------------

func TestLootCountsDifferentPowers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	loot := pushDiesCreatureForTest(g, me.ID, "Loot, the Nexus", rfcLootOracle, rfcLegendaryCreature, 2, 1)
	pushVanillaCreature(g, me.ID, "Two", 2, 2)
	pushVanillaCreature(g, me.ID, "Four", 4, 4)
	pushVanillaCreature(g, me.ID, "Four again", 4, 1)
	pushVanillaCreature(g, g.Seats[1].ID, "Theirs", 7, 7)
	spec, _ := Lookup(rfcLootOracle)
	if got := spec.ManaAbilities[0].ProducedFunc(g, me.ID, loot); got != OneColorOfAmount(2) {
		t.Errorf("produced = %q, want %q (powers 2 and 4 are two different powers)", got, OneColorOfAmount(2))
	}
}

func TestLootTapsForOneColourPick(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	loot := pushDiesCreatureForTest(g, me.ID, "Loot, the Nexus", rfcLootOracle, rfcLegendaryCreature, 2, 1)
	pushVanillaCreature(g, me.ID, "Four", 4, 4)
	if err := g.ActivateManaAbility(me.ID, loot, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
}

// --- Lyra, Archangel of Dawn --------------------------------------

func TestLyraArchangelOfDawnPumpsEachAngelOnLifegain(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	lyra := pushDiesCreatureForTest(g, me.ID, "Lyra, Archangel of Dawn", rfcLyraDawnOracle, "Legendary Creature — Angel Knight", 3, 3)
	angel := pushDiesCreatureForTest(g, me.ID, "Other Angel", "", "Creature — Angel", 2, 2)
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	theirs := pushDiesCreatureForTest(g, g.Seats[1].ID, "Their Angel", "", "Creature — Angel", 2, 2)
	rfcGainLife(g, me.ID, 4)
	passPriorityAroundTable(t, g)
	if rfcCountersOf(t, g, lyra, game.CounterPlusOne) != 1 || rfcCountersOf(t, g, angel, game.CounterPlusOne) != 1 {
		t.Error("each Angel I control gets a +1/+1 counter")
	}
	if rfcCountersOf(t, g, bear, game.CounterPlusOne) != 0 || rfcCountersOf(t, g, theirs, game.CounterPlusOne) != 0 {
		t.Error("a non-Angel or an opponent's Angel gets none")
	}
}

// --- Lyra, Tolarian Archangel -------------------------------------

func TestLyraTolarianMakesAnAngelAfterThreeDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Lyra, Tolarian Archangel", rfcLyraTolarianOracle, "Legendary Creature — Angel Wizard", 3, 3)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3-drawn) })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	angels := battlefieldIDsNamed(g, "Angel")
	if len(angels) != 1 {
		t.Fatalf("Angel tokens = %d, want 1", len(angels))
	}
	if effectivePower(t, g, angels[0]) != 3 || !hasAbility(effectiveAbilities(t, g, angels[0]), "flying") {
		t.Error("the Angel is a 3/3 flyer")
	}
}

func TestLyraTolarianMakesNothingAfterTwoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	pushDiesCreatureForTest(g, me.ID, "Lyra, Tolarian Archangel", rfcLyraTolarianOracle, "Legendary Creature — Angel Wizard", 3, 3)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2-drawn) })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Angel")); n != 0 {
		t.Errorf("Angel tokens = %d, want 0", n)
	}
}

func TestLyraTolarianActivationDrawsTwoOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := rfcSeats(g)
	toMain(t, g)
	lyra := pushDiesCreatureForTest(g, me.ID, "Lyra, Tolarian Archangel", rfcLyraTolarianOracle, "Legendary Creature — Angel Wizard", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, lyra, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	attackWith(t, g, opps[0].ID, lyra)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("cards drawn = %d, want 2", got)
	}
}

// --- Mabel, Valley Hero -------------------------------------------

func TestMabelPutsACounterOnACreatureThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := rfcSeats(g)
	old := pushDiesCreatureForTest(g, me.ID, "Old Bear", "", "Creature — Bear", 2, 2)
	mabel := castHoldCreature(t, g, "Mabel, Valley Hero", rfcLegendaryCreature, rfcMabelOracle, 1, 2)
	answerPickTarget(t, g, mabel)
	passPriorityAroundTable(t, g)
	if got := rfcCountersOf(t, g, mabel, game.CounterPlusOne); got != 1 {
		t.Errorf("Mabel counters = %d, want 1", got)
	}
	if got := rfcCountersOf(t, g, old, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature that did not enter this turn must not be a legal target, got %d", got)
	}
}
