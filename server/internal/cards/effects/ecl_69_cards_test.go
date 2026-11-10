package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_69_cards_test.go — Lorwyn Eclipsed, batch 69 (#4458).

const (
	ecl69AbigaleOracle   = "48a773d0-c433-4463-b50f-b6d4604a042c"
	ecl69BreOracle       = "59384481-3ca0-49dd-b94f-d3680b336f1e"
	ecl69SpewerOracle    = "77217a66-4ba8-443f-b915-178cfcc233e1"
	ecl69DoranOracle     = "dd35a377-0865-4b25-bd0f-aaead29a9f0e"
	ecl69HarvestOracle   = "70b9497f-0dd6-4b66-9675-57c7041fe3c0"
	ecl69FigureOracle    = "5d4ea910-3b16-4f98-89be-c1b2cfff4bf0"
	ecl69HurlerOracle    = "fb9cd9d0-8eea-4aa8-bb9c-bb0aac847421"
	ecl69LluwenOracle    = "81ff5764-72ce-4b9a-a14f-07413700d935"
	ecl69SkyswimmerOracl = "d709f396-021f-49eb-adb5-374aa790eaae"
	ecl69SneaklingOracle = "fe7a4361-561a-4bab-921e-d01bab19dad3"
)

func TestEcl69Batch69AreRegisteredFull(t *testing.T) {
	ecl67Registered(t, map[string]string{
		ecl69AbigaleOracle:   "Abigale, Eloquent First-Year",
		ecl69BreOracle:       "Bre of Clan Stoutarm",
		ecl69SpewerOracle:    "Chaos Spewer",
		ecl69DoranOracle:     "Doran, Besieged by Time",
		ecl69HarvestOracle:   "Dream Harvest",
		ecl69FigureOracle:    "Figure of Fable",
		ecl69HurlerOracle:    "Hovel Hurler",
		ecl69LluwenOracle:    "Lluwen, Imperfect Naturalist",
		ecl69SkyswimmerOracl: "Merrow Skyswimmer",
		ecl69SneaklingOracle: "Mischievous Sneakling",
	})
}

// --- Abigale -----------------------------------------------------------------

func TestEcl69AbigaleStripsAbilitiesThenGivesKeywordCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hasty", TypeLine: "Creature — Goblin",
		Power: 2, Toughness: 2, Keywords: []string{"haste", "deathtouch"}, Owner: opp.ID, Controller: opp.ID})
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Abigale, Eloquent First-Year", "Legendary Creature — Bird Bard", ecl69AbigaleOracle, 1, 1, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, victim, "haste") || hasEffectiveKeyword(t, g, victim, "deathtouch") {
		t.Error("the target kept its old abilities")
	}
	for _, kw := range []string{"flying", "first strike", "lifelink"} {
		if !hasEffectiveKeyword(t, g, victim, kw) {
			t.Errorf("the target has no %s from its counter", kw)
		}
	}
	_ = id
	// Indefinitely: it is still stripped the next turn.
	ecl6NextTurn(t, g)
	if hasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("the loss of abilities ended with the turn")
	}
}

// --- Bre ---------------------------------------------------------------------

func TestEcl69BrePutsAnExpensiveCardIntoHandAfterGainingLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Bre of Clan Stoutarm", "Legendary Creature — Giant Warrior", ecl69BreOracle, false)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	ids := mm2Library(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}"}, mm2Spell("Deep"))
	advanceTo(t, g, game.StepEnd)
	monSettle(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the nonland card (mana value 5 against 1 life gained) did not go to hand")
	}
	if zone := g.FindCardZoneForEffect(ids[0]); zone == nil || zone.Kind != game.ZoneExile {
		t.Error("the land was not exiled on the way")
	}
}

func TestEcl69BreDoesNothingWithoutLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Bre of Clan Stoutarm", "Legendary Creature — Giant Warrior", ecl69BreOracle, false)
	ids := mm2Library(me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{1}"}, mm2Spell("Deep"))
	advanceTo(t, g, game.StepEnd)
	monSettle(t, g)
	if zone := g.FindCardZoneForEffect(ids[0]); zone == nil || zone.Kind != game.ZoneLibrary {
		t.Error("the library was touched with no life gained")
	}
}

func TestEcl69BreGrantsFlyingAndLifelinkToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bre := pushCatalogPermanent(g, me.ID, "Bre of Clan Stoutarm", "Legendary Creature — Giant Warrior", ecl69BreOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "W", "C")
	if err := g.ActivateCatalogAbility(me.ID, bre, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(bre)}); err == nil {
		t.Error("Bre could target herself")
	}
	b16Activate(t, g, me.ID, bre, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") || !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the target lacks flying or lifelink")
	}
}

// --- Chaos Spewer ------------------------------------------------------------

func TestEcl69ChaosSpewerBlightsTwoUnlessYouPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me, _ := ecl6Seats(g)
		advanceToMain(t, g)
		id := ecl6Cast(t, g, me, "Chaos Spewer", "Creature — Goblin Warlock", ecl69SpewerOracle, 4, 4, game.CastSpellParams{})
		passPriorityAroundTable(t, g)
		if pay {
			b06AddMana(me, "C", "C")
		}
		answerPayUnless(t, g, me.ID, pay)
		monSettle(t, g)
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("pay=%v: the Spewer left the battlefield", pay)
		}
		want := 2
		if pay {
			want = 0
		}
		if got := c.Counters[game.CounterMinusOne]; got != want {
			t.Errorf("pay=%v: %d -1/-1 counters, want %d", pay, got, want)
		}
	}
}

// --- Doran -------------------------------------------------------------------

func TestEcl69DoranPumpsAnAttackerByTheDifferenceAndReducesCosts(t *testing.T) {
	spec, _ := Lookup(ecl69DoranOracle)
	if len(spec.CostModifiers) != 1 {
		t.Fatal("no cost modifier")
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Doran, Besieged by Time", "Legendary Creature — Treefolk Druid", ecl69DoranOracle, false)
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wall, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, wall); p != 4 || tt != 7 {
		t.Errorf("the attacker is %d/%d, want 4/7", p, tt)
	}
}

func TestEcl69DoranCostReductionFollowsToughnessOverPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Doran, Besieged by Time", "Legendary Creature — Treefolk Druid", ecl69DoranOracle, false)
	advanceToMain(t, g)
	for _, tc := range []struct {
		name         string
		power, tough int
		mana         int
		ok           bool
	}{{"toughness greater", 1, 4, 2, true}, {"equal", 2, 2, 2, false}, {"power greater", 4, 1, 2, false}} {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: "Creature — Wall", ManaCost: "{3}",
				Power: tc.power, Toughness: tc.tough, Owner: me.ID, Controller: me.ID})
			me.ManaPool = nil
		})
		b06AddMana(me, "C", "C")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if tc.ok && err != nil {
			t.Errorf("%s: a {3} spell with two mana: %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("%s: a {3} spell was cast for two mana", tc.name)
		}
		if err == nil {
			passPriorityAroundTable(t, g)
		}
	}
}

// --- Dream Harvest -----------------------------------------------------------

func TestEcl69DreamHarvestExilesUntilFiveAndGrantsAFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	advanceToMain(t, g)
	spell := func(n string) game.Card {
		return game.Card{Name: n, TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	}
	ids := mm2Library(opp, spell("One"), spell("Two"), spell("Three"), spell("Four"))
	ecl6Cast(t, g, me, "Dream Harvest", "Sorcery", ecl69HarvestOracle, 0, 0, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// Mana value 2 each: 2, 4, 6 — three cards.
	for i, id := range ids {
		zone := g.FindCardZoneForEffect(id)
		wantExile := i < 3
		if zone == nil || (zone.Kind == game.ZoneExile) != wantExile {
			t.Errorf("card %d: in %v, want exiled=%v", i, zone, wantExile)
		}
	}
	// The free cast: no mana in the pool.
	me.ManaPool = nil
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast of an exiled card: %v", err)
	}
}

// --- Figure of Fable ---------------------------------------------------------

func TestEcl69FigureOfFableGrowsThroughThreeStages(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	fig := pushCatalogPermanent(g, me.ID, "Figure of Fable", "Creature — Kithkin", ecl69FigureOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fig {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			}
		}
	})
	advanceToMain(t, g)
	act := func(idx int, mana ...string) {
		t.Helper()
		b06AddMana(me, mana...)
		b16Activate(t, g, me.ID, fig, idx, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
	}
	// The second stage does nothing before it is a Scout.
	act(1, "C", "G", "W")
	if p, tt := ptOf(t, g, fig); p != 1 || tt != 1 {
		t.Fatalf("a non-Scout stage 2 changed it to %d/%d", p, tt)
	}
	act(0, "G")
	c, _ := battlefieldCard(g, fig)
	if p, tt := ptOf(t, g, fig); p != 2 || tt != 3 || !c.HasSubtype("Scout") || !c.HasSubtype("Kithkin") {
		t.Fatalf("stage 1: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	act(1, "C", "G", "W")
	c, _ = battlefieldCard(g, fig)
	if p, tt := ptOf(t, g, fig); p != 4 || tt != 5 || !c.HasSubtype("Soldier") || c.HasSubtype("Scout") {
		t.Fatalf("stage 2: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	act(2, "C", "C", "C", "G", "W", "G")
	c, _ = battlefieldCard(g, fig)
	if p, tt := ptOf(t, g, fig); p != 7 || tt != 8 || !c.HasSubtype("Avatar") || c.HasSubtype("Soldier") {
		t.Fatalf("stage 3: %d/%d %v", p, tt, c.Effective().Subtypes)
	}
	// Protection from each of your opponents: their creature can't block it.
	_ = opp
	if !hasProtectionFromOpponents(c) {
		t.Error("no protection from each of your opponents")
	}
	// It lasts past the turn.
	ecl6NextTurn(t, g)
	if p, _ := ptOf(t, g, fig); p != 7 {
		t.Error("the effect ended with the turn")
	}
}

func hasProtectionFromOpponents(c game.Card) bool {
	for _, q := range game.ProtectionQualities(&c) {
		if q.Kind == game.ProtectionQualityOpponents {
			return true
		}
	}
	return false
}

// --- Hovel Hurler ------------------------------------------------------------

func TestEcl69HovelHurlerEntersWithCountersAndSpendsThemToGrantFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hurler := ecl6Cast(t, g, me, "Hovel Hurler", "Creature — Giant Warrior", ecl69HurlerOracle, 5, 5, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, hurler); c.Counters[game.CounterMinusOne] != 2 {
		t.Fatalf("%d -1/-1 counters on entering, want 2", c.Counters[game.CounterMinusOne])
	}
	b06AddMana(me, "R", "W")
	b16Activate(t, g, me.ID, hurler, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear), CounterKind: game.CounterMinusOne})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, hurler); c.Counters[game.CounterMinusOne] != 1 {
		t.Errorf("%d counters left, want one removed", c.Counters[game.CounterMinusOne])
	}
	if p, _ := ptOf(t, g, bear); p != 3 {
		t.Errorf("the Bear has %d power, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear has no flying")
	}
}

// --- Lluwen ------------------------------------------------------------------

func TestEcl69LluwenMillsFourAndMayPutACreatureOrLandOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	ids := mm2Library(me, mm2Spell("A"), game.Card{Name: "Elf", TypeLine: "Creature — Elf Scout"}, mm2Spell("B"), mm2Spell("C"), mm2Spell("Deep"))
	ecl6Cast(t, g, me, "Lluwen, Imperfect Naturalist", "Legendary Creature — Elf Druid", ecl69LluwenOracle, 1, 3, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pc := chooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no pick")
	}
	if len(pc.ChooseCards) != 1 || pc.ChooseCards[0] != ids[1] {
		t.Fatalf("offered %v, want only the creature card", pc.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if len(me.Library.Cards) == 0 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != ids[1] && me.Library.Cards[0].InstanceID != ids[1] {
		t.Error("the creature card is not at either end of the library")
	}
	if zone := g.FindCardZoneForEffect(ids[1]); zone == nil || zone.Kind != game.ZoneLibrary {
		t.Error("the creature card did not go back to the library")
	}
}

func TestEcl69LluwenMakesAWormForEachLandInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	lluwen := pushCatalogPermanent(g, me.ID, "Lluwen, Imperfect Naturalist", "Legendary Creature — Elf Druid", ecl69LluwenOracle, false)
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	})
	land := handCardFull(me, "Mountain", "Basic Land — Mountain", "", "", nil)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "B", "G", "B")
	b16Activate(t, g, me.ID, lluwen, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}})
	passPriorityAroundTable(t, g)
	// Two in the graveyard plus the discarded Mountain.
	if n, _ := countTokensNamed(g, me.ID, "Worm"); n != 3 {
		t.Errorf("%d Worm tokens, want 3", n)
	}
}

// --- Merrow Skyswimmer / Mischievous Sneakling -------------------------------

func TestEcl69SkyswimmerHasConvokeAndMakesAMerfolk(t *testing.T) {
	spec, _ := Lookup(ecl69SkyswimmerOracl)
	if spec.TapCost == nil {
		t.Fatal("no convoke")
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Merrow Skyswimmer", "Creature — Merfolk Soldier", ecl69SkyswimmerOracl, 3, 3, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Merfolk"); n != 1 {
		t.Errorf("%d Merfolk tokens, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") || !hasEffectiveKeyword(t, g, id, "vigilance") {
		t.Error("missing flying or vigilance")
	}
}

func TestEcl69SneaklingHasChangelingAndFlash(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Mischievous Sneakling", "Creature — Shapeshifter", ecl69SneaklingOracle, 2, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"changeling", "flash"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
	c, _ := battlefieldCard(g, id)
	if !c.HasSubtype("Merfolk") {
		t.Error("a changeling is every creature type")
	}
}
