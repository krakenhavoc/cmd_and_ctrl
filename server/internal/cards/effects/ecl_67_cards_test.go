package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_67_cards_test.go — Lorwyn Eclipsed, batches 67–70 (#4458). Each test
// plays the real catalog entry and asserts what its text says.

func ecl67Registered(t *testing.T, want map[string]string) {
	t.Helper()
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

const (
	ecl67LuminolluskOracle  = "c8bf2dce-7bcd-45cb-bf15-f53e24333a01"
	ecl67InformantOracle    = "b5058b5c-0359-42ec-ba9e-ae1371bb9c22"
	ecl67TillingOracle      = "79bc2c6d-b03d-4ceb-b468-293378ec73ae"
	ecl67CouncilOracle      = "9acfec0d-05a0-4b28-a883-e6863f1cb773"
	ecl67AdherentsOracle    = "d0fa43f6-d8f6-4674-b538-6d546b7dcf40"
	ecl67EyesOracle         = "c55a27cc-8e25-4714-b1da-86e5c1c6661a"
	ecl67FistsOracle        = "0d46a97a-b1d6-4bf3-b3ba-a5cdfd4c8bf5"
	ecl67PrismabasherOracle = "f75be4f2-d1a7-40c7-a0e4-ba8b03556ac6"
	ecl67UndercurrOracle    = "27a3aa64-1de7-4b0c-af23-52886e7d258a"
	ecl67PummelerOracle     = "ffa23c3a-76a8-414f-aa3b-fc09aabd78c5"
)

func TestEcl67Batch67AreRegisteredFull(t *testing.T) {
	ecl67Registered(t, map[string]string{
		ecl67LuminolluskOracle:  "Luminollusk",
		ecl67InformantOracle:    "Lys Alana Informant",
		ecl67TillingOracle:      "Midnight Tilling",
		ecl67CouncilOracle:      "Mistmeadow Council",
		ecl67AdherentsOracle:    "Moon-Vigil Adherents",
		ecl67EyesOracle:         "Morcant's Eyes",
		ecl67FistsOracle:        "Pitiless Fists",
		ecl67PrismabasherOracle: "Prismabasher",
		ecl67UndercurrOracle:    "Prismatic Undercurrents",
		ecl67PummelerOracle:     "Pummeler for Hire",
	})
}

// --- Luminollusk -------------------------------------------------------------

func TestEcl67LuminolluskGainsLifeForEachColorAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	b27Push(g, me.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	advanceToMain(t, g)
	before := lifeOf(g, me.ID)
	id := ecl6Cast(t, g, me, "Luminollusk", "Creature — Elemental", ecl67LuminolluskOracle, 3, 3, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// The test Luminollusk is a bare card with no colors: red and blue only.
	if got := lifeOf(g, me.ID) - before; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("no deathtouch")
	}
}

// --- Lys Alana Informant -----------------------------------------------------

func TestEcl67InformantSurveilsWhenItEntersAndWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	seedLibrary(me, "a", "b", "c")
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Lys Alana Informant", "Creature — Elf Scout", ecl67InformantOracle, 2, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("entering: no surveil prompt")
	}
	g.PendingChoices = nil
	b27Kill(g, id)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("dying: no surveil prompt")
	}
}

// --- Midnight Tilling --------------------------------------------------------

func TestEcl67MidnightTillingMillsFourAndMayReturnAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	toMainForCost(t, g)
	ids := mm2Library(me, mm2Spell("A"), game.Card{Name: "Relic", TypeLine: "Artifact"},
		mm2Spell("B"), mm2Spell("C"), mm2Spell("Deep"))
	id := handCardFull(me, "Midnight Tilling", "Instant", "", ecl67TillingOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the permanent card was not returned to hand")
	}
	if me.Library.Size() != 1 {
		t.Errorf("library %d, want exactly four milled", me.Library.Size())
	}
}

// --- Mistmeadow Council ------------------------------------------------------

func TestEcl67MistmeadowCouncilIsCheaperWithAKithkinAndDraws(t *testing.T) {
	spec, _ := Lookup(ecl67CouncilOracle)
	if len(spec.SelfCostModifiers) != 1 {
		t.Fatal("no self cost modifier")
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	seedLibrary(me, "a", "b")
	advanceToMain(t, g)
	before := handSizeOf(g, me)
	ecl6Cast(t, g, me, "Mistmeadow Council", "Creature — Kithkin Advisor", ecl67CouncilOracle, 4, 5, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// The card was put in (+1), cast (-1) and the draw added one.
	if got := handSizeOf(g, me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

// --- Moon-Vigil Adherents ----------------------------------------------------

func TestEcl67MoonVigilAdherentsCountsCreaturesAndGraveyardCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	id := pushCatalogPermanent(g, me.ID, "Moon-Vigil Adherents", "Creature — Elf Druid", ecl67AdherentsOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 0, 0
			}
		}
	})
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	// Two creatures on the battlefield and one creature card: +3/+3 over 0/0.
	if p, tt := ptOf(t, g, id); p != 3 || tt != 3 {
		t.Errorf("Adherents are %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample")
	}
}

// --- Morcant's Eyes ----------------------------------------------------------

func TestEcl67MorcantsEyesSurveilsAtUpkeepAndSacrificesForElves(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	seedLibrary(me, "a", "b", "c")
	eyes := pushCatalogPermanent(g, me.ID, "Morcant's Eyes", "Kindred Enchantment — Elf", ecl67EyesOracle, false)
	g.WithWriteLock(func() {
		for _, n := range []string{"Elf One", "Elf Two"} {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	advanceToMain(t, g)
	b06AddMana(me, "G", "G", "C", "C", "C", "C")
	b16Activate(t, g, me.ID, eyes, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Morcant's Eyes")) != 0 {
		t.Error("the Eyes were not sacrificed")
	}
	// Two Elf cards plus the sacrificed Eyes themselves.
	if n, _ := countTokensNamed(g, me.ID, "Elf"); n != 3 {
		t.Errorf("%d Elf tokens, want 3", n)
	}
}

func TestEcl67MorcantsEyesSurveilsOnYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Morcant's Eyes", "Kindred Enchantment — Elf", ecl67EyesOracle, false)
	seedLibrary(me, "a", "b", "c")
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil prompt at your upkeep")
	}
}

// --- Pitiless Fists ----------------------------------------------------------

func TestEcl67PitilessFistsPumpsAndFightsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Pitiless Fists", "Enchantment — Aura", ecl67FistsOracle, 0, 0,
		game.CastSpellParams{Targets: cardRefs(mine)})
	ecl6ResolveSpell(t, g)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
	if len(battlefieldIDsNamed(g, "Theirs")) != 0 {
		t.Error("the opposing creature survived the 4-power fight")
	}
	if len(battlefieldIDsNamed(g, "Mine")) != 1 {
		t.Error("the enchanted creature died to 3 damage on a 4/4")
	}
}

func TestEcl67PitilessFistsRefusesACreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Pitiless Fists", TypeLine: "Enchantment — Aura",
			OracleID: ecl67FistsOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(theirs)}); err == nil {
		t.Error("enchanting an opponent's creature was allowed")
	}
}

// --- Prismabasher ------------------------------------------------------------

func TestEcl67PrismabasherPumpsUpToXCreaturesByX(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	a := b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	b := b27Push(g, me.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	c := b27Push(g, me.ID, "Green Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Prismabasher", "Creature — Elemental", ecl67PrismabasherOracle, 7, 7, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no Prismabasher target prompt")
	}
	// Red, blue and green: X is 3, so all three bears may be chosen.
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if p, tt := ptOf(t, g, id); p != 5 || tt != 5 {
			t.Errorf("a target is %d/%d, want 5/5", p, tt)
		}
	}
	ecl6NextTurn(t, g)
	if p, _ := ptOf(t, g, a); p != 2 {
		t.Error("the boost outlasted the turn")
	}
}

// --- Prismatic Undercurrents -------------------------------------------------

func TestEcl67PrismaticUndercurrentsFetchesBasicsAndAllowsAnExtraLand(t *testing.T) {
	spec, _ := Lookup(ecl67UndercurrOracle)
	if spec.AdditionalLandPlays != 1 {
		t.Errorf("additional land plays %d, want 1", spec.AdditionalLandPlays)
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	b27Push(g, me.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := 0; i < 4; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
				Owner: me.ID, Controller: me.ID})
		}
	})
	advanceToMain(t, g)
	before := handSizeOf(g, me)
	ecl6Cast(t, g, me, "Prismatic Undercurrents", "Enchantment", ecl67UndercurrOracle, 0, 0, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pc := searchChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no search prompt")
	}
	answerSearchByID(t, g, me.ID, pc.SearchCards[:2]...)
	passPriorityAroundTable(t, g)
	// Red and blue among my permanents: up to two basic lands.
	if got := handSizeOf(g, me) - before; got != 2 {
		t.Errorf("hand grew by %d, want 2 basic lands", got)
	}
}

// --- Pummeler for Hire -------------------------------------------------------

func TestEcl67PummelerGainsLifeEqualToTheGreatestGiantPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	b12Creature(g, me.ID, "Big Giant", "Creature — Giant", 6, 6)
	b12Creature(g, me.ID, "Small Giant", "Creature — Giant", 2, 2)
	b12Creature(g, me.ID, "Huge Bear", "Creature — Bear", 9, 9)
	advanceToMain(t, g)
	before := lifeOf(g, me.ID)
	id := ecl6Cast(t, g, me, "Pummeler for Hire", "Creature — Giant Mercenary", ecl67PummelerOracle, 4, 4, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID) - before; got != 6 {
		t.Errorf("gained %d life, want 6", got)
	}
	for _, kw := range []string{"reach", "vigilance"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}
