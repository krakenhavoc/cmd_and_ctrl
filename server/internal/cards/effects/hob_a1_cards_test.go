package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a1_cards_test.go — The Hobbit / The Hobbit Eternal, first batch
// (#4454): ten cards whose seams all exist. Each test casts or plays the
// real catalog entry and asserts what the card's text says.

const (
	hobA1BilbosDeadlySliceOracle = "4c054431-3667-4e28-9d68-03b43447c35b"
	hobA1BanishingLightOracle    = "f28b21a6-f7ce-437a-8c5b-0423cb55cefb"
	hobA1OrdinaryBearOracle      = "89a2fab8-b074-42aa-a871-d9c5de2d0895"
	hobA1TrollOfKhazadDumOracle  = "222fd9fc-7af4-4269-bbb9-cd7b8ac029bd"
	hobA1GaladhrimGuideOracle    = "9204de14-cbb6-40e5-9dcc-44cce3723e8e"
	hobA1WillowWindOracle        = "ee3fd92c-48e0-4651-aeda-7eb2c75acbb0"
	hobA1CaptainOfUmbarOracle    = "1941f993-41d7-49e3-8891-41b3cecd5a9d"
	hobA1HithlainKnotsOracle     = "2c80fc55-86a0-403e-82ce-09ee03095d8a"
	hobA1SmaugsFuryOracle        = "5c034266-8e63-4b43-8f9e-564e6fff6ec1"
	hobA1DoriOracle              = "58acd6b2-730c-4472-a34a-f2784977aca8"
)

func TestHobA1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA1BilbosDeadlySliceOracle: "Bilbo's Deadly Slice",
		hobA1BanishingLightOracle:    "Banishing Light",
		hobA1OrdinaryBearOracle:      "Ordinary Bear",
		hobA1TrollOfKhazadDumOracle:  "Troll of Khazad-dûm",
		hobA1GaladhrimGuideOracle:    "Galadhrim Guide",
		hobA1WillowWindOracle:        "Willow-Wind",
		hobA1CaptainOfUmbarOracle:    "Captain of Umbar",
		hobA1HithlainKnotsOracle:     "Hithlain Knots",
		hobA1SmaugsFuryOracle:        "Smaug's Fury",
		hobA1DoriOracle:              "Dori, Bearer of Friends",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

func TestHobA1BilbosDeadlySliceDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Bilbo's Deadly Slice", "Instant", hobA1BilbosDeadlySliceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

func TestHobA1BanishingLightExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	light := castCatalogSpell(t, g, "Banishing Light", "Enchantment", hobA1BanishingLightOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(light) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after Banishing Light left, %d copies under its owner's control, want 1", n)
	}
}

// A land is not a legal pick, and neither is your own permanent: the
// trigger's clause is "nonland permanent an opponent controls".
func TestHobA1BanishingLightOnlyTargetsOpponentsNonlandPermanents(t *testing.T) {
	spec, _ := Lookup(hobA1BanishingLightOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

func TestHobA1OrdinaryBearIsARegisteredVanillaBody(t *testing.T) {
	spec, ok := Lookup(hobA1OrdinaryBearOracle)
	if !ok {
		t.Fatal("Ordinary Bear is not registered")
	}
	if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("a vanilla body declares no abilities")
	}
}

func TestHobA1TrollOfKhazadDumNeedsThreeBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	troll := b12Push(g, me.ID, "Troll of Khazad-dûm", "Creature — Troll", hobA1TrollOfKhazadDumOracle, 6, 5)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	brAttack(t, g, troll)

	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: troll}, {Blocker: b, Attacker: troll},
	}), game.BlockReasonTooFewBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: troll}, {Blocker: b, Attacker: troll}, {Blocker: c, Attacker: troll},
	}); err != nil {
		t.Fatalf("three blockers are legal: %v", err)
	}
}

func TestHobA1TrollOfKhazadDumSwampcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := pushLibraryCardForTest(me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	pushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	cid, p := cycleFromHand(t, g, "Troll of Khazad-dûm", "Creature — Troll", hobA1TrollOfKhazadDumOracle, "{C}")
	if !p.Graveyard.Contains(cid) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, p.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != swamp {
		t.Fatalf("want a search offering only the Swamp, got %+v", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, p.ID, []uuid.UUID{swamp}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !p.Hand.Contains(swamp) {
		t.Error("swampcycling did not fetch the Swamp")
	}
}

// hobA1ScryOnEntry casts the creature and answers the ETB scry, keeping
// every card on top.
func hobA1ScryOnEntry(t *testing.T, name, typeLine, oracle string, want int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three")
	id := castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s should have resolved onto the battlefield", name)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt from the ETB: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != want {
		t.Fatalf("scry %d should offer %d cards, got %d", want, want, len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	return g, id
}

func TestHobA1GaladhrimGuideScriesTwoWhenItEnters(t *testing.T) {
	hobA1ScryOnEntry(t, "Galadhrim Guide", "Creature — Elf Scout", hobA1GaladhrimGuideOracle, 2)
}

func TestHobA1WillowWindFliesAndScriesTwo(t *testing.T) {
	g, id := hobA1ScryOnEntry(t, "Willow-Wind", "Creature — Elemental", hobA1WillowWindOracle, 2)
	if !effectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("Willow-Wind should have flying")
	}
}

func TestHobA1CaptainOfUmbarLootsForOneAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	captain := pushCatalogPermanent(g, me.ID, "Captain of Umbar", "Creature — Human Pirate", hobA1CaptainOfUmbarOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, captain, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if !b20Tapped(t, g, captain) {
		t.Error("the loot has a tap cost")
	}
}

func TestHobA1HithlainKnotsTapsScriesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "One", "Two", "Three")
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	castCatalogSpell(t, g, "Hithlain Knots", "Instant", hobA1HithlainKnotsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if !b20Tapped(t, g, victim) {
		t.Error("the target creature should be tapped")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a scry 1 prompt, got %+v", c)
	}
	if me.Hand.Size() != hand {
		t.Error("the draw happened before the scry was answered (CR 608.2c)")
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d: the draw follows the scry", got, hand+1)
	}
}

func TestHobA1SmaugsFuryPumpsAndGrantsReachAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushVanillaCreature(g, me.ID, "Recipient", 2, 2)
	castCatalogSpell(t, g, "Smaug's Fury", "Instant", hobA1SmaugsFuryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, id)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 2 {
		t.Errorf("got %d/%d, want 5/2", c.CurrentPower(), c.CurrentToughness())
	}
	for _, kw := range []string{"reach", "first strike"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("missing %s", kw)
		}
	}
}

func TestHobA1DoriHasTrampleAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Dori, Bearer of Friends", "Legendary Creature — Dwarf Warrior", hobA1DoriOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Dori should have resolved")
	}
	if !effectiveAbilitiesContain(t, g, id, "trample") {
		t.Error("Dori should have trample")
	}
	if n := namedOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("Treasure tokens = %d, want 1", n)
	}
}
