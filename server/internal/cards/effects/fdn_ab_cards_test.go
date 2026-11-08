package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_ab_cards_test.go — Foundations, the cards from A and B that batch 1
// (fdn_b1_cards_test.go) left (#4465). Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	fdnABIndrikOracle     = "88aa032d-dc11-4cae-81f3-ce66353963e0"
	fdnABWolfOracle       = "4c48de92-e053-4af7-b936-d7ec0875e0d3"
	fdnABDragonOracle     = "60b03a44-a7bd-48fa-8c8f-1704aed02cd7"
	fdnABVitalityOracle   = "635fe908-6f58-4c9e-ac55-0775a7c6f278"
	fdnABStomperOracle    = "6a381121-9134-489c-8c7e-44be1936b0e2"
	fdnABArbiterOracle    = "540693f3-985c-4a4a-945c-c957c5aad395"
	fdnABEpiphanyOracle   = "fe1556cc-32d0-4e51-bb15-f3fce669a5e1"
	fdnABArchwayOracle    = "f2cf087a-d89a-4145-9394-945ce41c8cca"
	fdnABCavalryOracle    = "6a92b011-fd01-4b6a-a17c-73ac1967ff98"
	fdnABBerserkerOracle  = "396cc8ce-9efa-4fc7-90a6-b439b2b021db"
	fdnABRangerOracle     = "db794af5-ba2e-4b10-b213-e481e8a50432"
	fdnABShriekmassOracle = "ce45d6bd-74e3-401b-983d-028d3fb942da"
	fdnABCollectorOracle  = "93088b0c-d158-473a-9c06-aa2f344a9868"
	fdnABBountyOracle     = "63b0538e-aa94-4e23-adf1-704cbf5bc3e5"
)

func TestFdnABEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnABIndrikOracle:     "Affectionate Indrik",
		fdnABWolfOracle:       "Ambush Wolf",
		fdnABDragonOracle:     "Ancestor Dragon",
		fdnABVitalityOracle:   "Angel of Vitality",
		fdnABStomperOracle:    "Apothecary Stomper",
		fdnABArbiterOracle:    "Arbiter of Woe",
		fdnABEpiphanyOracle:   "Arcane Epiphany",
		fdnABArchwayOracle:    "Archway Angel",
		fdnABCavalryOracle:    "Axgard Cavalry",
		fdnABBerserkerOracle:  "Battlesong Berserker",
		fdnABRangerOracle:     "Beast-Kin Ranger",
		fdnABShriekmassOracle: "Billowing Shriekmass",
		fdnABCollectorOracle:  "Bloodtithe Collector",
		fdnABBountyOracle:     "Brass's Bounty",
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

func TestFdnABPrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Ancestor Dragon", "Creature — Dragon", fdnABDragonOracle, 5, 6, []string{"flying"}},
		{"Angel of Vitality", "Creature — Angel", fdnABVitalityOracle, 2, 2, []string{"flying"}},
		{"Apothecary Stomper", "Creature — Elephant", fdnABStomperOracle, 4, 4, []string{"vigilance"}},
		{"Arbiter of Woe", "Creature — Demon", fdnABArbiterOracle, 5, 4, []string{"flying"}},
		{"Archway Angel", "Creature — Angel", fdnABArchwayOracle, 3, 4, []string{"flying"}},
		{"Beast-Kin Ranger", "Creature — Elf Ranger", fdnABRangerOracle, 3, 3, []string{"trample"}},
		{"Billowing Shriekmass", "Creature — Spirit", fdnABShriekmassOracle, 2, 3, []string{"flying"}},
		{"Bloodtithe Collector", "Creature — Vampire Noble", fdnABCollectorOracle, 3, 4, []string{"flying"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
	spec, _ := Lookup(fdnABWolfOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Ambush Wolf keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Affectionate Indrik ------------------------------------------------------

func TestFdnABAffectionateIndrikMayFightOnEntry(t *testing.T) {
	for _, yes := range []bool{true, false} {
		name := "declined"
		if yes {
			name = "fights"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			indrik := castHoldCreature(t, g, "Affectionate Indrik", "Creature — Beast", fdnABIndrikOracle, 4, 4)
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, victim)
			answerLatestTriggerPrompt(t, g, me.ID, yes)
			passPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if got := g.Battlefield.Contains(victim); got == yes {
				t.Errorf("the 2/2 is on the battlefield: %v", got)
			}
			want := 0
			if yes {
				want = 2
			}
			if got := damageMarkedOn(g, indrik); got != want {
				t.Errorf("the Indrik has %d damage, want %d", got, want)
			}
		})
	}
}

func TestFdnABAffectionateIndrikCannotFightYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Affectionate Indrik", "Creature — Beast", fdnABIndrikOracle, 4, 4)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, mine) {
		t.Error("a creature you control is offered as the target")
	}
}

// --- Ambush Wolf --------------------------------------------------------------

func TestFdnABAmbushWolfExilesACardFromAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := graveyardCardOf(g, opp, "Their Card", "Instant", "{R}", "")
	castHoldCreature(t, g, "Ambush Wolf", "Creature — Wolf", fdnABWolfOracle, 4, 2)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("the chosen graveyard card was not exiled")
	}
}

// --- Ancestor Dragon ----------------------------------------------------------

func TestFdnABAncestorDragonGainsOneForEachAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ancestor Dragon", "Creature — Dragon", fdnABDragonOracle, 5, 6)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	advanceToMain(t, g)
	life := me.Life
	// The Dragon stays home: "one or more creatures you control attack".
	declareAttack(t, g, opp.ID, a, b, c)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want one trigger gaining 3", life, me.Life)
	}
}

// --- Angel of Vitality --------------------------------------------------------

func TestFdnABAngelOfVitalityAddsOneToEachGainAndIsBigAtTwentyFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	angel := b12Push(g, me.ID, "Angel of Vitality", "Creature — Angel", fdnABVitalityOracle, 2, 2)
	if me.Life < 25 {
		t.Fatalf("the table starts at %d life", me.Life)
	}
	if p, tough := effectivePTOf(t, g, angel); p != 4 || tough != 4 {
		t.Errorf("at %d life the Angel is %d/%d, want 4/4", me.Life, p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 24-me.Life) })
	if me.Life != 24 {
		t.Fatalf("life %d, want 24 (a loss is not replaced)", me.Life)
	}
	if p, tough := effectivePTOf(t, g, angel); p != 2 || tough != 2 {
		t.Errorf("at 24 life the Angel is %d/%d, want 2/2", p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	if me.Life != 27 {
		t.Errorf("gaining 2 from 24: life %d, want 27", me.Life)
	}
	if p, tough := effectivePTOf(t, g, angel); p != 4 || tough != 4 {
		t.Errorf("at %d life the Angel is %d/%d, want 4/4", me.Life, p, tough)
	}
}

// --- Apothecary Stomper -------------------------------------------------------

func TestFdnABApothecaryStomperModes(t *testing.T) {
	t.Run("counters", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		castAndResolveCreature(t, g, "Apothecary Stomper", "Creature — Elephant", fdnABStomperOracle)
		c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		if n := plusOneCounters(g, bear); n != 2 {
			t.Errorf("the Bear has %d +1/+1 counters, want 2", n)
		}
	})
	t.Run("life", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		life := me.Life
		castAndResolveCreature(t, g, "Apothecary Stomper", "Creature — Elephant", fdnABStomperOracle)
		c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		passPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d -> %d, want +4", life, me.Life)
		}
	})
}

// --- Arbiter of Woe -----------------------------------------------------------

func TestFdnABArbiterOfWoeCostsACreatureAndDrainsTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if _, err := tryCastWithSacrifice(g, "Arbiter of Woe", "Creature — Demon", fdnABArbiterOracle, nil); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	seedLibrary(me, "Drawn")
	for _, p := range g.Seats[1:] {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	lives := b27Lives(g)
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = handSize(p)
	}
	castWithSacrifice(t, g, "Arbiter of Woe", "Creature — Demon", fdnABArbiterOracle, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the sacrifice is paid while casting")
	}
	hand := handSize(me)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		discardFromHand(t, g, p.ID)
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 2
		if i == 0 {
			want = lives[i] + 2
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
		if i > 0 && handSize(p) != hands[i]-1 {
			t.Errorf("seat %d hand %d -> %d, want one card discarded", i, hands[i], handSize(p))
		}
	}
	if handSize(me) != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, handSize(me))
	}
}

// --- Arcane Epiphany ----------------------------------------------------------

func TestFdnABArcaneEpiphanyCostsOneLessWithAWizardAndDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, fdnABEpiphanyOracle, "Instant", "{3}{U}{U}"); got != 5 {
		t.Errorf("with no Wizard: %d, want 5", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Wizard", TypeLine: "Creature — Human Wizard"})
	if got := selfPricedMV(t, g, me, fdnABEpiphanyOracle, "Instant", "{3}{U}{U}"); got != 4 {
		t.Errorf("with a Wizard: %d, want 4", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three", "Four")
	castCatalogSpell(t, g, "Arcane Epiphany", "Instant", fdnABEpiphanyOracle, nil)
	hand := handSize(me)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+3 {
		t.Errorf("hand %d -> %d, want three cards drawn", hand, handSize(me))
	}
}

// --- Archway Angel ------------------------------------------------------------

func TestFdnABArchwayAngelGainsTwoForEachGateYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLandFor(g, me.ID, "Azorius Guildgate", "Land — Gate")
	pushLandFor(g, me.ID, "Boros Guildgate", "Land — Gate")
	pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	pushLandFor(g, opp.ID, "Their Gate", "Land — Gate")
	life := me.Life
	castAndResolveCreature(t, g, "Archway Angel", "Creature — Angel", fdnABArchwayOracle)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 for two Gates", life, me.Life)
	}
}

// --- Axgard Cavalry -----------------------------------------------------------

func TestFdnABAxgardCavalryGivesHasteForATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cavalry := b12Push(g, me.ID, "Axgard Cavalry", "Creature — Dwarf Berserker", fdnABCavalryOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, cavalry, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
	if !isTapped(g, cavalry) {
		t.Error("the Cavalry is not tapped")
	}
}

// --- Battlesong Berserker -----------------------------------------------------

func TestFdnABBattlesongBerserkerPumpsAndGrantsMenaceWhenYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := b12Push(g, me.ID, "Battlesong Berserker", "Creature — Human Berserker", fdnABBerserkerOracle, 3, 4)
	attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The Berserker stays home: "whenever you attack".
	declareAttack(t, g, opp.ID, attacker)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if !hasID(pick.PickTargetCards, berserker) || hasID(pick.PickTargetCards, theirs) {
		t.Errorf("legal set %v: want creatures you control only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, attacker)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, attacker); p != 3 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, attacker, "menace") {
		t.Error("the attacker did not gain menace")
	}
}

// --- Beast-Kin Ranger ---------------------------------------------------------

func TestFdnABBeastKinRangerGrowsWhenAnotherCreatureOfYoursEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ranger := b12Push(g, me.ID, "Beast-Kin Ranger", "Creature — Elf Ranger", fdnABRangerOracle, 3, 3)
	advanceToMain(t, g)

	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	emitETBForTest(g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, ranger); p != 3 {
		t.Errorf("an opponent's creature entering: power %d, want 3", p)
	}

	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	emitETBForTest(g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, ranger); p != 4 || tough != 3 {
		t.Errorf("after your creature entered the Ranger is %d/%d, want 4/3", p, tough)
	}
}

// --- Billowing Shriekmass -----------------------------------------------------

func TestFdnABBillowingShriekmassMillsThreeAndHasThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D", "E")
	for me.Graveyard.Size() < 4 {
		me.Graveyard.PushTop(game.NewCard("Old", me.ID))
	}
	if n := me.Graveyard.Size(); n != 4 {
		t.Fatalf("graveyard starts at %d, want 4", n)
	}
	held := b12Push(g, me.ID, "Billowing Shriekmass", "Creature — Spirit", fdnABShriekmassOracle, 2, 3)
	if p, tough := effectivePTOf(t, g, held); p != 2 || tough != 3 {
		t.Errorf("with four cards in the graveyard the Shriekmass is %d/%d, want 2/3", p, tough)
	}
	shriek := castHoldCreature(t, g, "Billowing Shriekmass", "Creature — Spirit", fdnABShriekmassOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if n := me.Graveyard.Size(); n != 7 {
		t.Fatalf("graveyard %d, want 7 after milling three", n)
	}
	for _, id := range []uuid.UUID{held, shriek} {
		if p, tough := effectivePTOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("with seven cards in the graveyard a Shriekmass is %d/%d, want 4/4", p, tough)
		}
	}
}

// --- Bloodtithe Collector -----------------------------------------------------

func TestFdnABBloodtitheCollectorNeedsAnOpponentToHaveLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		name := "no life lost"
		if lost {
			name = "an opponent lost life"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			for _, p := range g.Seats[1:] {
				p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
			}
			advanceToMain(t, g)
			hands := make(map[uuid.UUID]int, len(g.Seats))
			for _, p := range g.Seats {
				hands[p.ID] = handSize(p)
			}
			if lost {
				g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
			}
			castAndResolveCreature(t, g, "Bloodtithe Collector", "Creature — Vampire Noble", fdnABCollectorOracle)
			passPriorityAroundTable(t, g)
			for _, p := range g.Seats[1:] {
				if got := discardChoiceFor(g, p.ID) != nil; got != lost {
					t.Errorf("seat %s asked to discard: %v, want %v", p.ID, got, lost)
				}
				if lost {
					discardFromHand(t, g, p.ID)
					if handSize(p) != hands[p.ID]-1 {
						t.Errorf("seat %s hand %d -> %d, want one card discarded", p.ID, hands[p.ID], handSize(p))
					}
				}
			}
		})
	}
}

// --- Brass's Bounty -----------------------------------------------------------

func TestFdnABBrasssBountyMakesATreasurePerLandYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		pushLand(g, me.ID, "Island")
	}
	pushLand(g, opp.ID, "Their Island")
	castCatalogSpell(t, g, "Brass's Bounty", "Sorcery", fdnABBountyOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Treasure", me.ID); n != 3 {
		t.Errorf("%d Treasures, want 3", n)
	}
}
