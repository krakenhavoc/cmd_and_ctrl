package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b37_test.go — Assassin's Creed cards of S82 GAP batch 37 (#5948).
// Each test plays the real catalog entry.

const (
	gapb37HarlequinOracle = "10e0ccdf-9a31-402a-9fbf-f501ce63fa52"
	gapb37MentorOracle    = "60a3345c-3234-4f4b-baab-df12dad6b44b"
	gapb37RoshanOracle    = "f60b21c1-d192-4a02-9b1f-a08e9e08a84a"
	gapb37AvelineOracle   = "5c8c4d53-f364-40da-989b-cfe1e3f79b7d"
	gapb37CleopatraOracle = "e7a80464-1522-48d7-bf82-f9894f5c3ae0"
	gapb37EzioOracle      = "3017ac79-37f1-4cc1-8efd-4add027618a2"
	gapb37ArnoOracle      = "ae90daea-6644-4235-bca6-b94ee9925f63"
	gapb37EdwardOracle    = "b2fe6d17-d045-45c6-9ca2-21e73d2e38d2"
	gapb37AltairOracle    = "43dc4b0c-e437-4abc-a3fb-48bd3098de88"
	gapb37AchillesOracle  = "b921a672-721e-4bd8-a10b-a6de8c13c7d8"
	gapb37AdewaleOracle   = "e30d9e08-cf88-4e2d-95e1-7144d8fea5c7"
)

func TestGapB37CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb37HarlequinOracle: "Merciless Harlequin",
		gapb37MentorOracle:    "Poison-Blade Mentor",
		gapb37RoshanOracle:    "Roshan, Hidden Magister",
		gapb37AvelineOracle:   "Aveline de Grandpré",
		gapb37CleopatraOracle: "Cleopatra, Exiled Pharaoh",
		gapb37EzioOracle:      "Ezio Auditore da Firenze",
		gapb37ArnoOracle:      "Arno Dorian",
		gapb37EdwardOracle:    "Edward Kenway",
		gapb37AltairOracle:    "Altaïr Ibn-La'Ahad",
		gapb37AchillesOracle:  "Achilles Davenport",
		gapb37AdewaleOracle:   "Adéwalé, Breaker of Chains",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// gapb37ConnectThenPostcombat attacks seat 1 with a creature of `typeLine`
// under seat 0 and walks on to the postcombat main phase.
func gapb37ConnectThenPostcombat(t *testing.T, g *game.Game, typeLine string) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	dealer := effectstest.PushDiesCreatureForTest(g, me.ID, "Dealer", "", typeLine, 2, 2)
	effectstest.AttackWith(t, g, opp.ID, dealer)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
}

func gapb37Hand(p *game.Player, name, typeLine, cost, oracle string, colors []string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Colors: colors, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- Merciless Harlequin ------------------------------------------------------

func TestGapB37HarlequinFreerunsThenDrawsAndLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb37ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	id := gapb37Hand(me, "Merciless Harlequin", "Creature — Human Assassin", "{2}{B}", gapb37HarlequinOracle, []string{"B"}, 2, 1)
	effectstest.FloatForTest(g, me, "CB")
	hand, life := me.Hand.Size(), me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning"}); err != nil {
		t.Fatalf("freerunning {1}{B} was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Merciless Harlequin did not enter")
	}
	if got := me.Hand.Size(); got != hand-1+1 {
		t.Errorf("hand %d -> %d, want the spell gone and one card drawn", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
}

func TestGapB37HarlequinNeedsAnAssassinToConnectForFreerunning(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb37ConnectThenPostcombat(t, g, "Creature — Human Warrior")
	id := gapb37Hand(me, "Merciless Harlequin", "Creature — Human Assassin", "{2}{B}", gapb37HarlequinOracle, []string{"B"}, 2, 1)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning"}); err == nil {
		t.Error("freerunning was accepted with no Assassin or commander connecting")
	}
}

// --- Poison-Blade Mentor ------------------------------------------------------

func TestGapB37MentorGivesAnotherAssassinDeathtouchWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mentor := effectstest.B12Push(g, me.ID, "Poison-Blade Mentor", "Creature — Human Assassin", gapb37MentorOracle, 2, 1)
	if !effectstest.EffectiveAbilitiesContain(t, g, mentor, "deathtouch") {
		t.Error("the Mentor has no deathtouch")
	}
	assassin := effectstest.PushDiesCreatureForTest(g, me.ID, "Other Assassin", "", "Creature — Human Assassin", 2, 2)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mentor, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, assassin)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, assassin, "deathtouch") {
		t.Error("the other Assassin did not gain deathtouch")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("a non-Assassin gained deathtouch")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, assassin, "deathtouch") {
		t.Error("the deathtouch outlasted the turn")
	}
}

// --- Achilles Davenport and Arno Dorian (Assassin lords) ----------------------

func TestGapB37AchillesPumpsOnlyYourOtherAssassins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	achilles := effectstest.B12Push(g, me.ID, "Achilles Davenport", "Legendary Creature — Human Assassin", gapb37AchillesOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Assassin", "Creature — Human Assassin", 2, 2)
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my Assassin is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, achilles); p != 3 || tt != 3 {
		t.Errorf("Achilles is %d/%d, want 3/3 (other Assassins only)", p, tt)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("a Bear got the bonus: power %d", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Assassin got the bonus: power %d", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, achilles, "menace") {
		t.Error("Achilles has no menace")
	}
}

func TestGapB37AchillesFreerunsForUAndB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb37ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	id := gapb37Hand(me, "Achilles Davenport", "Legendary Creature — Human Assassin", "{2}{U}{B}", gapb37AchillesOracle, []string{"U", "B"}, 3, 3)
	effectstest.FloatForTest(g, me, "UB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning"}); err != nil {
		t.Fatalf("freerunning {U}{B} was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Achilles did not enter")
	}
}

func TestGapB37ArnoPumpsOtherAssassinsPowerAndDisguises(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arno := effectstest.B12Push(g, me.ID, "Arno Dorian", "Legendary Creature — Human Assassin", gapb37ArnoOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Assassin", "Creature — Human Assassin", 2, 2)
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 4 || tt != 2 {
		t.Errorf("my Assassin is %d/%d, want 4/2", p, tt)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, arno); p != 3 {
		t.Errorf("Arno's own power is %d, want 3", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, arno, "deathtouch") {
		t.Error("Arno has no deathtouch")
	}
	// Disguise {B}{R}: cast face down for {3} as a 2/2 with ward {2}.
	defs, ok := effects.Lookup(gapb37ArnoOracle)
	if !ok || len(defs.AlternativeCosts) != 1 || defs.AlternativeCosts[0].Key != "disguise" {
		t.Fatalf("Arno has no disguise offer: %+v", defs.AlternativeCosts)
	}
}

// --- Aveline de Grandpré ------------------------------------------------------

func TestGapB37AvelineGrowsADeathtouchCreatureByTheDamageItDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Aveline de Grandpré", "Legendary Creature — Human Assassin", gapb37AvelineOracle, 3, 3)
	toucher := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Snake", TypeLine: "Creature — Snake", Power: 2, Toughness: 2, Keywords: []string{"deathtouch"}})
	plain := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, toucher, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, toucher, game.CounterPlusOne); got != 2 {
		t.Errorf("the deathtouch creature has %d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CounterOn(g, plain, game.CounterPlusOne); got != 0 {
		t.Errorf("the plain creature has %d +1/+1 counters, want 0", got)
	}
}

// --- Ezio Auditore da Firenze -------------------------------------------------

func TestGapB37EzioGrantsAssassinSpellsFreerunningBB(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           bool
	}{
		{"Assassin spell", "Creature — Human Assassin", true},
		{"non-Assassin spell", "Creature — Human Warrior", false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Ezio Auditore da Firenze", "Legendary Creature — Human Assassin", gapb37EzioOracle, 3, 2)
		gapb37ConnectThenPostcombat(t, g, "Creature — Human Assassin")
		id := gapb37Hand(me, "Test Spell", tc.typeLine, "{4}{R}", "", []string{"R"}, 3, 3)
		effectstest.FloatForTest(g, me, "BB")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFreerunning})
		if tc.want && err != nil {
			t.Errorf("%s: freerunning {B}{B} was refused: %v", tc.name, err)
		}
		if !tc.want && err == nil {
			t.Errorf("%s: freerunning {B}{B} was accepted", tc.name)
		}
	}
}

func TestGapB37EzioFreerunningNeedsAnAssassinToConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Ezio Auditore da Firenze", "Legendary Creature — Human Assassin", gapb37EzioOracle, 3, 2)
	gapb37ConnectThenPostcombat(t, g, "Creature — Human Warrior")
	id := gapb37Hand(me, "Test Assassin", "Creature — Human Assassin", "{4}{R}", "", []string{"R"}, 3, 3)
	effectstest.FloatForTest(g, me, "BB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFreerunning}); err == nil {
		t.Error("granted freerunning was accepted with no Assassin or commander connecting")
	}
}

func gapb37EzioHit(t *testing.T, oppLife int, pay bool) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ezio := effectstest.B12Push(g, me.ID, "Ezio Auditore da Firenze", "Legendary Creature — Human Assassin", gapb37EzioOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, oppLife-opp.Life) })
	effectstest.AttackWith(t, g, opp.ID, ezio)
	effectstest.PassPriorityAroundTable(t, g)
	if oppLife-3 <= 10 {
		effectstest.FloatForTest(g, me, "WUBRG")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
	}
	return g, opp
}

func TestGapB37EzioPaidWUBRGMakesALowLifePlayerLose(t *testing.T) {
	_, opp := gapb37EzioHit(t, 13, true)
	if !opp.Eliminated {
		t.Errorf("the player at %d life did not lose the game", opp.Life)
	}
}

func TestGapB37EzioDeclinedPaymentLeavesThePlayerInTheGame(t *testing.T) {
	_, opp := gapb37EzioHit(t, 13, false)
	if opp.Eliminated {
		t.Error("the player lost without Ezio's controller paying")
	}
}

func TestGapB37EzioAsksNothingWhenThePlayerHasMoreThanTenLife(t *testing.T) {
	g, opp := gapb37EzioHit(t, 40, true)
	if effectstest.HasPayUnlessFor(g, g.Seats[0].ID) {
		t.Error("Ezio offered the payment against a player with more than 10 life")
	}
	if opp.Eliminated || opp.Life != 37 {
		t.Errorf("opponent: eliminated=%v life=%d, want 37 and alive", opp.Eliminated, opp.Life)
	}
}

// --- Cleopatra, Exiled Pharaoh ------------------------------------------------

func TestGapB37CleopatraAlliesCountersTwoOtherLegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cleo := effectstest.B12Push(g, me.ID, "Cleopatra, Exiled Pharaoh", "Legendary Creature — Human Noble", gapb37CleopatraOracle, 2, 4)
	a := effectstest.B12Creature(g, me.ID, "Legend A", "Legendary Creature — Human", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Legend B", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no Allies target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, cleo) || effectstest.HasID(p.PickTargetCards, plain) {
		t.Error("Cleopatra or a nonlegendary creature was offered as a target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, a, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, b, game.CounterPlusOne) != 1 {
		t.Error("Allies did not put a +1/+1 counter on each target")
	}
	if effectstest.CountersOn(g, cleo, game.CounterPlusOne) != 0 {
		t.Error("Cleopatra put a counter on herself")
	}
}

func TestGapB37CleopatraBetrayalDrawsPerCounterAndCostsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Cleopatra, Exiled Pharaoh", "Legendary Creature — Human Noble", gapb37CleopatraOracle, 2, 4)
	legend := effectstest.B12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(legend, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(legend, "charge", 1)
	})
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legend) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("drew %d cards, want 3 (one per counter of any kind)", got-hand)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
}

func TestGapB37CleopatraBetrayalIgnoresNonlegendaryAndCounterlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Cleopatra, Exiled Pharaoh", "Legendary Creature — Human Noble", gapb37CleopatraOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare Legend", "Legendary Creature — Human", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2) })
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(bare)
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d, want no change", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Edward Kenway ------------------------------------------------------------

func gapb37Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

func gapb37Treasures(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Treasure" {
			n++
		}
	}
	return n
}

func TestGapB37EdwardMakesATreasureForEachTappedAssassinPirateOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	edward := effectstest.B12Push(g, me.ID, "Edward Kenway", "Legendary Creature — Human Assassin Pirate", gapb37EdwardOracle, 5, 5)
	assassin := effectstest.B12Creature(g, me.ID, "Assassin", "Creature — Human Assassin", 1, 1)
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 1, 1)
	vehicle := effectstest.B12Permanent(g, me.ID, "Skiff", "Artifact — Vehicle")
	idle := effectstest.B12Creature(g, me.ID, "Untapped Pirate", "Creature — Human Pirate", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pirate", "Creature — Human Pirate", 1, 1)
	for _, id := range []uuid.UUID{edward, assassin, pirate, vehicle, bear, theirs} {
		gapb37Tap(g, id)
	}
	_ = idle
	before := gapb37Treasures(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	// Edward (Assassin and Pirate, counted once), the Assassin, the Pirate and
	// the Vehicle: four. The Bear, the untapped Pirate and their Pirate don't.
	if got := gapb37Treasures(g, me.ID) - before; got != 4 {
		t.Errorf("created %d Treasures, want 4", got)
	}
}

func TestGapB37EdwardExilesTheTopCardWhenAVehicleConnects(t *testing.T) {
	for _, vehicle := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Edward Kenway", "Legendary Creature — Human Assassin Pirate", gapb37EdwardOracle, 5, 5)
		typeLine := "Creature — Bear"
		if vehicle {
			typeLine = "Artifact Creature — Vehicle"
		}
		dealer := effectstest.B12Creature(g, me.ID, "Dealer", typeLine, 3, 3)
		topCard, _ := opp.Library.Top()
		size := opp.Library.Size()
		effectstest.AttackWith(t, g, opp.ID, dealer)
		effectstest.PassPriorityAroundTable(t, g)
		z := g.FindCardZoneForEffect(topCard.InstanceID)
		if !vehicle {
			if opp.Library.Size() != size || z == nil || z.Kind != game.ZoneLibrary {
				t.Errorf("a non-Vehicle connecting exiled a card")
			}
			continue
		}
		if opp.Library.Size() != size-1 || z == nil || z.Kind != game.ZoneExile {
			t.Fatalf("library %d -> %d, top card zone %v, want one card exiled", size, opp.Library.Size(), z != nil && z.Kind == game.ZoneExile)
		}
	}
}

// --- Adéwalé, Breaker of Chains -----------------------------------------------

func TestGapB37AdewaleTakesAnAssassinPirateOrVehicleFromTheTopSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mk := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	filler := make([]uuid.UUID, 0, 4)
	for i := 0; i < 4; i++ {
		filler = append(filler, mk("Filler", "Creature — Bear"))
	}
	skiff := mk("Skiff", "Artifact — Vehicle")
	rogue := mk("Rogue", "Creature — Human Assassin")
	deep := mk("Deep Pirate", "Creature — Human Pirate") // below the top six
	_ = deep
	// Order: the last pushed is on top, so the top six are rogue, skiff and four fillers.
	effectstest.CastHoldCreature(t, g, "Adéwalé, Breaker of Chains", "Legendary Creature — Human Assassin Pirate", gapb37AdewaleOracle, 4, 1)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no card pick was offered")
	}
	if !effectstest.HasID(pick.ChooseCards, rogue) || !effectstest.HasID(pick.ChooseCards, skiff) {
		t.Errorf("the Assassin and the Vehicle were not both offered: %v", pick.ChooseCards)
	}
	for _, id := range filler {
		if effectstest.HasID(pick.ChooseCards, id) {
			t.Error("a non-matching card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{skiff}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(skiff); z == nil || z.Kind != game.ZoneHand {
		t.Errorf("the Vehicle is in %v, want the hand", z)
	}
	if z := g.FindCardZoneForEffect(rogue); z == nil || z.Kind != game.ZoneLibrary {
		t.Errorf("the unchosen Assassin is in %v, want the library (bottom)", z)
	}
}

func TestGapB37AdewaleReturnsFromTheGraveyardWhenAVehicleConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adewale := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: adewale, Name: "Adéwalé, Breaker of Chains", OracleID: gapb37AdewaleOracle,
		TypeLine: "Legendary Creature — Human Assassin Pirate", Power: 4, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	skiff := effectstest.B12Creature(g, me.ID, "Skiff", "Artifact Creature — Vehicle", 3, 3)
	effectstest.AttackWith(t, g, opp.ID, skiff)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(adewale); z == nil || z.Kind != game.ZoneHand {
		t.Errorf("Adéwalé is in %v, want the hand", z)
	}
}

func TestGapB37AdewaleStaysPutWhenANonVehicleConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adewale := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: adewale, Name: "Adéwalé, Breaker of Chains", OracleID: gapb37AdewaleOracle,
		TypeLine: "Legendary Creature — Human Assassin Pirate", Power: 4, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(adewale); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("Adéwalé is in %v, want the graveyard", z)
	}
}

// --- Roshan, Hidden Magister --------------------------------------------------

func TestGapB37RoshanMakesYourOtherCreaturesAssassins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roshan := effectstest.B12Push(g, me.ID, "Roshan, Hidden Magister", "Legendary Creature — Human Assassin", gapb37RoshanOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	subs := effectstest.EffectiveSubtypes(t, g, bear)
	if !effects.HasFold(subs, "Assassin") || !effects.HasFold(subs, "Bear") {
		t.Errorf("my Bear is %v, want Bear and Assassin", subs)
	}
	if effects.HasFold(effectstest.EffectiveSubtypes(t, g, theirs), "Assassin") {
		t.Error("an opponent's creature became an Assassin")
	}
	if got := effectstest.EffectiveSubtypes(t, g, roshan); len(got) != 2 {
		t.Errorf("Roshan's types are %v, want Human Assassin unchanged", got)
	}
}

func TestGapB37RoshanMakesCreatureCardsInHandAssassins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Roshan, Hidden Magister", "Legendary Creature — Human Assassin", gapb37RoshanOracle, 4, 4)
	inHand := gapb37Hand(me, "Hand Bear", "Creature — Bear", "{1}{G}", "", []string{"G"}, 2, 2)
	land := gapb37Hand(me, "Hand Forest", "Basic Land — Forest", "", "", nil, 0, 0)
	if !effects.HasFold(gapb37HandSubtypes(t, g, inHand), "Assassin") {
		t.Error("a creature card in hand is not an Assassin")
	}
	if effects.HasFold(gapb37HandSubtypes(t, g, land), "Assassin") {
		t.Error("a land card in hand became an Assassin")
	}
}

func TestGapB37RoshanGivesFaceDownCreaturesMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Roshan, Hidden Magister", "Legendary Creature — Human Assassin", gapb37RoshanOracle, 4, 4)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hidden Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		FaceDown: true, FaceDownKind: game.FaceDownDisguised, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hidden Ogre", TypeLine: "Creature — Ogre", Power: 2, Toughness: 2,
		FaceDown: true, FaceDownKind: game.FaceDownDisguised, Owner: opp.ID, Controller: opp.ID,
	})
	faceUp := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "menace") {
		t.Error("my face-down creature has no menace")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "menace") || effectstest.EffectiveAbilitiesContain(t, g, faceUp, "menace") {
		t.Error("an opponent's face-down or my face-up creature got menace")
	}
}

func TestGapB37RoshanDrawsWhenYourPermanentIsTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Roshan, Hidden Magister", "Legendary Creature — Human Assassin", gapb37RoshanOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventTurnedFaceUp, CardID: theirs, Actor: opp.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatal("an opponent's permanent turning face up triggered Roshan")
	}
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventTurnedFaceUp, CardID: mine, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d, want +1 card and -1 life", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- Altaïr Ibn-La'Ahad -------------------------------------------------------

func TestGapB37AltairExilesAnAssassinWithAMemoryCounterAndCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altair := effectstest.B12Push(g, me.ID, "Altaïr Ibn-La'Ahad", "Legendary Creature — Human Assassin", gapb37AltairOracle, 3, 3)
	if !effectstest.EffectiveAbilitiesContain(t, g, altair, "first strike") {
		t.Error("Altaïr has no first strike")
	}
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Fallen Assassin", TypeLine: "Creature — Human Assassin", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: bear, Name: "Fallen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(altair, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the graveyard card")
	}
	if effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("a non-Assassin card was offered")
	}
	effectstest.AnswerPickTarget(t, g, dead)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	c, ok := g.LookupCardForEffect(dead)
	z := g.FindCardZoneForEffect(dead)
	if !ok || z == nil || z.Kind != game.ZoneExile || c.Counters["memory"] != 1 {
		t.Fatalf("the Assassin is in %v with counters %v, want exile with a memory counter", z, c.Counters)
	}
	copies := 0
	for _, bc := range g.Battlefield.Cards {
		if bc.IsToken() && bc.Name == "Fallen Assassin" {
			copies++
			if !bc.Tapped || bc.AttackingTarget == uuid.Nil {
				t.Errorf("the copy is tapped=%v attacking=%v, want tapped and attacking", bc.Tapped, bc.AttackingTarget != uuid.Nil)
			}
		}
	}
	if copies != 1 {
		t.Fatalf("%d copy tokens, want 1", copies)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, bc := range g.Battlefield.Cards {
		if bc.IsToken() && bc.Name == "Fallen Assassin" {
			t.Error("the copy was not exiled at end of combat")
		}
	}
}

// gapb37HandSubtypes is a card's subtypes with the layers applied, wherever it is.
func gapb37HandSubtypes(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("card %s not found", id)
		}
		out = c.Effective().Subtypes
	})
	return out
}
