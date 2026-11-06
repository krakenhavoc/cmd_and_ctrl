package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_mp_cards_test.go — The Lost Caverns of Ixalan, cards M–P (#4470).
// Each test plays the real catalog entry and asserts what its text says.

const (
	lciMPBrawlerOracle     = "06905dfb-b2df-4849-8848-fd3c964aa8c9"
	lciMPVeteranOracle     = "25a9c864-59c6-4230-a234-bf78bf1ef24c"
	lciMPScribeOracle      = "8978a30a-7918-4851-92ef-f40fdf90c379"
	lciMPBrinefangOracle   = "cbbd507e-06da-4b47-befc-61db8c92539e"
	lciMPDraughtOracle     = "05da6b06-fb5f-41e8-b0a2-d949646e36d9"
	lciMPMightOracle       = "5e4107ce-8450-4501-9cb3-a6344ef320f1"
	lciMPGuidewingOracle   = "eb161d0b-e534-463c-b1d1-7210bfba5d28"
	lciMPSpiderOracle      = "e5aa32b6-2eaf-4b48-81cc-cd3db976513d"
	lciMPBristlebackOracle = "d33509d6-9d0c-4248-bb6d-0a9b7712f3a0"
	lciMPCloudGuardOracle  = "922a8e27-a367-40c6-84df-86ed32f71c7c"
	lciMPPuzzleDoorOracle  = "6bcf21ef-6641-4752-a0da-937a9277c812"
	lciMPAxejawOracle      = "54452df0-560c-4561-9dd2-3903ae49c91b"
)

func TestLciMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciMPBrawlerOracle:     "Malamet Brawler",
		lciMPVeteranOracle:     "Malamet Veteran",
		lciMPScribeOracle:      "Malamet War Scribe",
		lciMPBrinefangOracle:   "Marauding Brinefang",
		lciMPDraughtOracle:     "Mephitic Draught",
		lciMPMightOracle:       "Might of the Ancestors",
		lciMPGuidewingOracle:   "Miner's Guidewing",
		lciMPSpiderOracle:      "Mineshaft Spider",
		lciMPBristlebackOracle: "Nurturing Bristleback",
		lciMPCloudGuardOracle:  "Oltec Cloud Guard",
		lciMPPuzzleDoorOracle:  "Orazca Puzzle-Door",
		lciMPAxejawOracle:      "Pathfinding Axejaw",
	} {
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

// --- Malamet Brawler -------------------------------------------------------------

func TestLciMPMalametBrawlerGivesAnAttackerTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawler := b12Push(g, me.ID, "Malamet Brawler", "Creature — Cat Warrior", lciMPBrawlerOracle, 2, 2)
	other := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "My Idle Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, brawler, other)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, idle) || !hasID(pick.PickTargetCards, brawler) || !hasID(pick.PickTargetCards, other) {
		t.Errorf("offered %v; want the attacking creatures only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, other, "trample") {
		t.Error("the targeted attacker did not gain trample")
	}
	if hasEffectiveKeyword(t, g, brawler, "trample") {
		t.Error("the Brawler gained trample; only the target does")
	}
}

// --- Malamet Veteran -------------------------------------------------------------

func TestLciMPMalametVeteranDescendFourPutsACounterOnTargetCreature(t *testing.T) {
	setup := func(t *testing.T, permanents int) (g *game.Game, me *game.Player, veteran, bear uuid.UUID) {
		g = newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		veteran = b12Push(g, me.ID, "Malamet Veteran", "Creature — Cat Warrior", lciMPVeteranOracle, 5, 4)
		bear = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		lciSTPermanentCardsInGraveyard(me, permanents)
		// A nonpermanent card is not counted (CR 110.4).
		pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
		if !hasEffectiveKeyword(t, g, veteran, "trample") {
			t.Error("Malamet Veteran has no trample")
		}
		declareAttack(t, g, opp.ID, veteran)
		return g, me, veteran, bear
	}
	t.Run("four permanent cards", func(t *testing.T) {
		g, me, veteran, bear := setup(t, 4)
		pick := latestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatal("no target prompt with four permanent cards in the graveyard")
		}
		// Any creature: the non-attacking bear is a legal target.
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
			t.Errorf("the bear has %d +1/+1 counters, want 1", got)
		}
		if got := countersOn(g, veteran, game.CounterPlusOne); got != 0 {
			t.Errorf("the Veteran has %d +1/+1 counters, want 0", got)
		}
	})
	t.Run("three permanent cards", func(t *testing.T) {
		g, me, _, bear := setup(t, 3)
		if latestPickTarget(g, me.ID) != nil {
			t.Fatal("the descend trigger asked for a target with three permanent cards in the graveyard")
		}
		passPriorityAroundTable(t, g)
		if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
			t.Errorf("the bear has %d +1/+1 counters, want 0", got)
		}
	})
}

// --- Malamet War Scribe ----------------------------------------------------------

func TestLciMPMalametWarScribePumpsCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	scribe := msc17Play(t, g, me, game.Card{Name: "Malamet War Scribe", TypeLine: "Creature — Cat Warrior", OracleID: lciMPScribeOracle, Power: 4, Toughness: 3})
	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 4 || tough != 3 {
		t.Errorf("your bear is %d/%d, want 4/3", p, tough)
	}
	if p, tough := effectivePower(t, g, scribe), effectiveToughness(t, g, scribe); p != 6 || tough != 4 {
		t.Errorf("the Scribe is %d/%d, want 6/4 (it pumps itself too)", p, tough)
	}
	if p, tough := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Marauding Brinefang ---------------------------------------------------------

func TestLciMPMaraudingBrinefangHasWardThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fang := b12Push(g, me.ID, "Marauding Brinefang", "Creature — Dinosaur", lciMPBrinefangOracle, 6, 7)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, fang)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	var cost string
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == opp.ID {
			cost = c.PayCost
		}
	}
	if cost != "{3}" {
		t.Errorf("an opponent targeting the Brinefang is asked to pay %q, want ward {3}", cost)
	}
}

func TestLciMPMaraudingBrinefangIslandcyclesForAnIslandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Breeding Pool", "Land — Forest Island"),
		searchTestLand("Swamp", "Basic Land — Swamp"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id := lciSTActivateFromHand(t, g, "Marauding Brinefang", "Creature — Dinosaur", lciMPBrinefangOracle, "{C}{C}", 0)
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Brinefang is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after islandcycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Island-typed lands", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the fetched land did not reach the hand")
	}
}

// --- Mephitic Draught ------------------------------------------------------------

func TestLciMPMephiticDraughtDrawsOnEntryAndWhenItGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life
	draught := msc17Play(t, g, me, game.Card{Name: "Mephitic Draught", TypeLine: "Artifact", OracleID: lciMPDraughtOracle})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards on entry, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d after entry, want %d", me.Life, life-1)
	}
	hand, life = me.Hand.Size(), me.Life
	killCreature(t, g, me.ID, draught)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards when it was put into the graveyard, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d after it was put into the graveyard, want %d", me.Life, life-1)
	}
}

func TestLciMPMephiticDraughtDoesNotTriggerWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	draught := b12Push(g, me.ID, "Mephitic Draught", "Artifact", lciMPDraughtOracle, 0, 0)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	exileCreature(t, g, draught)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards when it was exiled, want 0", got)
	}
}

// --- Might of the Ancestors ------------------------------------------------------

func TestLciMPMightOfTheAncestorsPumpsAndGrantsVigilanceAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Might of the Ancestors", "Enchantment", lciMPMightOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Fatal("the bear did not gain vigilance")
	}
	declareAttack(t, g, opp.ID, bear)
	if tappedNow(g, bear) {
		t.Error("vigilance: attacking tapped the bear")
	}
}

// --- Miner's Guidewing -----------------------------------------------------------

func TestLciMPMinersGuidewingDiesAndATargetExplores(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wing := b12Push(g, me.ID, "Miner's Guidewing", "Creature — Bird", lciMPGuidewingOracle, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, wing, kw) {
			t.Errorf("Miner's Guidewing has no %s", kw)
		}
	}
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	killCreature(t, g, me.ID, wing)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the dies trigger did not ask for a target")
	}
	if hasID(pick.PickTargetCards, wing) {
		t.Error("the dead Guidewing was offered as the explorer")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed card was not put into the graveyard")
	}
}

// --- Mineshaft Spider ------------------------------------------------------------

func TestLciMPMineshaftSpiderMayMillTwo(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		library := me.Library.Size()
		spider := castHoldCreature(t, g, "Mineshaft Spider", "Creature — Spider", lciMPSpiderOracle, 3, 4)
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := 0
		if accept {
			want = 2
		}
		if got := library - me.Library.Size(); got != want {
			t.Errorf("accept=%v: milled %d cards, want %d", accept, got, want)
		}
		if !hasEffectiveKeyword(t, g, spider, "reach") {
			t.Error("Mineshaft Spider has no reach")
		}
	}
}

// --- Nurturing Bristleback -------------------------------------------------------

func TestLciMPNurturingBristlebackMakesAThreeThreeDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{Name: "Nurturing Bristleback", TypeLine: "Creature — Dinosaur", OracleID: lciMPBristlebackOracle, Power: 5, Toughness: 5})
	dinos := trekB33Tokens(g, me.ID, "Dinosaur")
	if len(dinos) != 1 {
		t.Fatalf("%d Dinosaur tokens, want 1", len(dinos))
	}
	c := layeredCard(t, g, dinos[0].InstanceID)
	if !c.IsCreature() || c.Power != 3 || c.Toughness != 3 || len(c.Colors) != 1 || c.Colors[0] != "G" || !c.HasSubtype("Dinosaur") {
		t.Errorf("token = %s %d/%d %v, want a 3/3 green Dinosaur", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
	if hasEffectiveKeyword(t, g, c.InstanceID, "trample") {
		t.Error("the Dinosaur token has trample; the printed token has no abilities")
	}
}

func TestLciMPNurturingBristlebackForestcyclesForAForestCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Temple Garden", "Land — Forest Plains"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
	)
	lciSTActivateFromHand(t, g, "Nurturing Bristleback", "Creature — Dinosaur", lciMPBristlebackOracle, "{C}{C}", 0)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after forestcycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Forest-typed lands", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the fetched Forest did not reach the hand")
	}
}

// --- Oltec Cloud Guard -----------------------------------------------------------

func TestLciMPOltecCloudGuardMakesAGnomeAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guard := msc17Play(t, g, me, game.Card{Name: "Oltec Cloud Guard", TypeLine: "Creature — Human Soldier", OracleID: lciMPCloudGuardOracle, Power: 3, Toughness: 2})
	gnomes := trekB33Tokens(g, me.ID, "Gnome")
	if len(gnomes) != 1 {
		t.Fatalf("%d Gnome tokens, want 1", len(gnomes))
	}
	if c := layeredCard(t, g, gnomes[0].InstanceID); !c.IsCreature() || !c.IsArtifact() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("token = %s %d/%d %v, want a 1/1 colorless Gnome artifact creature", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
	if !hasEffectiveKeyword(t, g, guard, "flying") {
		t.Error("Oltec Cloud Guard has no flying")
	}
}

// --- Orazca Puzzle-Door ----------------------------------------------------------

func TestLciMPOrazcaPuzzleDoorTakesOneOfTheTopTwoAndBinsTheOther(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	door := pushCatalogPermanent(g, me.ID, "Orazca Puzzle-Door", "Artifact", lciMPPuzzleDoorOracle, false)
	advanceToMain(t, g)
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, door, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana for {1}")
	}
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, door, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(door) {
		t.Error("the door was not sacrificed as the cost")
	}
	lib := me.Library.Size()
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the top two cards")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lower}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(lower) {
		t.Error("the chosen card is not in hand")
	}
	if !me.Graveyard.Contains(upper) {
		t.Error("the other of the top two should be put into the graveyard")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d -> %d, want exactly the top two gone", lib, me.Library.Size())
	}
}

// --- Pathfinding Axejaw ----------------------------------------------------------

func TestLciMPPathfindingAxejawExploresOnEntry(t *testing.T) {
	t.Run("a land goes to the hand", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		top := putOnTopForTest(me, "Forest", "Basic Land — Forest")
		axejaw := msc17Play(t, g, me, game.Card{Name: "Pathfinding Axejaw", TypeLine: "Creature — Dinosaur", OracleID: lciMPAxejawOracle, Power: 4, Toughness: 3})
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(top) {
			t.Error("the revealed land did not go to the hand")
		}
		if got := countersOn(g, axejaw, game.CounterPlusOne); got != 0 {
			t.Errorf("the Axejaw has %d +1/+1 counters after revealing a land, want 0", got)
		}
	})
	t.Run("a nonland gives a counter and may stay on top", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
		axejaw := msc17Play(t, g, me, game.Card{Name: "Pathfinding Axejaw", TypeLine: "Creature — Dinosaur", OracleID: lciMPAxejawOracle, Power: 4, Toughness: 3})
		passPriorityAroundTable(t, g)
		if got := countersOn(g, axejaw, game.CounterPlusOne); got != 1 {
			t.Errorf("the Axejaw has %d +1/+1 counters, want 1", got)
		}
		answerExploreQuestion(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
		if !me.Library.Contains(top) || me.Graveyard.Contains(top) {
			t.Error("the revealed card left the library after the player kept it on top")
		}
	})
}
