package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		effectstest.AssertKeywords(t, g, id, c.keywords...)
	}
	spec, _ := effects.Lookup(fdnABWolfOracle)
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			indrik := effectstest.CastHoldCreature(t, g, "Affectionate Indrik", "Creature — Beast", fdnABIndrikOracle, 4, 4)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, victim)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if got := g.Battlefield.Contains(victim); got == yes {
				t.Errorf("the 2/2 is on the battlefield: %v", got)
			}
			want := 0
			if yes {
				want = 2
			}
			if got := effectstest.DamageMarkedOn(g, indrik); got != want {
				t.Errorf("the Indrik has %d damage, want %d", got, want)
			}
		})
	}
}

func TestFdnABAffectionateIndrikCannotFightYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Affectionate Indrik", "Creature — Beast", fdnABIndrikOracle, 4, 4)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("a creature you control is offered as the target")
	}
}

// --- Ambush Wolf --------------------------------------------------------------

func TestFdnABAmbushWolfExilesACardFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Card", "Instant", "{R}", "")
	effectstest.CastHoldCreature(t, g, "Ambush Wolf", "Creature — Wolf", fdnABWolfOracle, 4, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("the chosen graveyard card was not exiled")
	}
}

// --- Ancestor Dragon ----------------------------------------------------------

func TestFdnABAncestorDragonGainsOneForEachAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ancestor Dragon", "Creature — Dragon", fdnABDragonOracle, 5, 6)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	c := effectstest.PushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	// The Dragon stays home: "one or more creatures you control attack".
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want one trigger gaining 3", life, me.Life)
	}
}

// --- Angel of Vitality --------------------------------------------------------

func TestFdnABAngelOfVitalityAddsOneToEachGainAndIsBigAtTwentyFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Angel of Vitality", "Creature — Angel", fdnABVitalityOracle, 2, 2)
	if me.Life < 25 {
		t.Fatalf("the table starts at %d life", me.Life)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, angel); p != 4 || tough != 4 {
		t.Errorf("at %d life the Angel is %d/%d, want 4/4", me.Life, p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 24-me.Life) })
	if me.Life != 24 {
		t.Fatalf("life %d, want 24 (a loss is not replaced)", me.Life)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, angel); p != 2 || tough != 2 {
		t.Errorf("at 24 life the Angel is %d/%d, want 2/2", p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	if me.Life != 27 {
		t.Errorf("gaining 2 from 24: life %d, want 27", me.Life)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, angel); p != 4 || tough != 4 {
		t.Errorf("at %d life the Angel is %d/%d, want 4/4", me.Life, p, tough)
	}
}

// --- Apothecary Stomper -------------------------------------------------------

func TestFdnABApothecaryStomperModes(t *testing.T) {
	t.Run("counters", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		effectstest.CastAndResolveCreature(t, g, "Apothecary Stomper", "Creature — Elephant", fdnABStomperOracle)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.PlusOneCounters(g, bear); n != 2 {
			t.Errorf("the Bear has %d +1/+1 counters, want 2", n)
		}
	})
	t.Run("life", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Apothecary Stomper", "Creature — Elephant", fdnABStomperOracle)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d -> %d, want +4", life, me.Life)
		}
	})
}

// --- Arbiter of Woe -----------------------------------------------------------

func TestFdnABArbiterOfWoeCostsACreatureAndDrainsTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if _, err := effectstest.TryCastWithSacrifice(g, "Arbiter of Woe", "Creature — Demon", fdnABArbiterOracle, nil); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	effectstest.SeedLibrary(me, "Drawn")
	for _, p := range g.Seats[1:] {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	lives := effectstest.B27Lives(g)
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = effectstest.HandSize(p)
	}
	effectstest.CastWithSacrifice(t, g, "Arbiter of Woe", "Creature — Demon", fdnABArbiterOracle, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the sacrifice is paid while casting")
	}
	hand := effectstest.HandSize(me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 2
		if i == 0 {
			want = lives[i] + 2
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
		if i > 0 && effectstest.HandSize(p) != hands[i]-1 {
			t.Errorf("seat %d hand %d -> %d, want one card discarded", i, hands[i], effectstest.HandSize(p))
		}
	}
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, effectstest.HandSize(me))
	}
}

// --- Arcane Epiphany ----------------------------------------------------------

func TestFdnABArcaneEpiphanyCostsOneLessWithAWizardAndDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, fdnABEpiphanyOracle, "Instant", "{3}{U}{U}"); got != 5 {
		t.Errorf("with no Wizard: %d, want 5", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Wizard", TypeLine: "Creature — Human Wizard"})
	if got := effectstest.SelfPricedMV(t, g, me, fdnABEpiphanyOracle, "Instant", "{3}{U}{U}"); got != 4 {
		t.Errorf("with a Wizard: %d, want 4", got)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.CastCatalogSpell(t, g, "Arcane Epiphany", "Instant", fdnABEpiphanyOracle, nil)
	hand := effectstest.HandSize(me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+3 {
		t.Errorf("hand %d -> %d, want three cards drawn", hand, effectstest.HandSize(me))
	}
}

// --- Archway Angel ------------------------------------------------------------

func TestFdnABArchwayAngelGainsTwoForEachGateYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, me.ID, "Azorius Guildgate", "Land — Gate")
	effectstest.PushLandFor(g, me.ID, "Boros Guildgate", "Land — Gate")
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, opp.ID, "Their Gate", "Land — Gate")
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Archway Angel", "Creature — Angel", fdnABArchwayOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 for two Gates", life, me.Life)
	}
}

// --- Axgard Cavalry -----------------------------------------------------------

func TestFdnABAxgardCavalryGivesHasteForATap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cavalry := effectstest.B12Push(g, me.ID, "Axgard Cavalry", "Creature — Dwarf Berserker", fdnABCavalryOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, cavalry, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
	if !effectstest.IsTapped(g, cavalry) {
		t.Error("the Cavalry is not tapped")
	}
}

// --- Battlesong Berserker -----------------------------------------------------

func TestFdnABBattlesongBerserkerPumpsAndGrantsMenaceWhenYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	berserker := effectstest.B12Push(g, me.ID, "Battlesong Berserker", "Creature — Human Berserker", fdnABBerserkerOracle, 3, 4)
	attacker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The Berserker stays home: "whenever you attack".
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if !effectstest.HasID(pick.PickTargetCards, berserker) || effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Errorf("legal set %v: want creatures you control only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 3 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, attacker, "menace") {
		t.Error("the attacker did not gain menace")
	}
}

// --- Beast-Kin Ranger ---------------------------------------------------------

func TestFdnABBeastKinRangerGrowsWhenAnotherCreatureOfYoursEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ranger := effectstest.B12Push(g, me.ID, "Beast-Kin Ranger", "Creature — Elf Ranger", fdnABRangerOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)

	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EmitETBForTest(g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, ranger); p != 3 {
		t.Errorf("an opponent's creature entering: power %d, want 3", p)
	}

	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.EmitETBForTest(g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, ranger); p != 4 || tough != 3 {
		t.Errorf("after your creature entered the Ranger is %d/%d, want 4/3", p, tough)
	}
}

// --- Billowing Shriekmass -----------------------------------------------------

func TestFdnABBillowingShriekmassMillsThreeAndHasThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	for me.Graveyard.Size() < 4 {
		me.Graveyard.PushTop(game.NewCard("Old", me.ID))
	}
	if n := me.Graveyard.Size(); n != 4 {
		t.Fatalf("graveyard starts at %d, want 4", n)
	}
	held := effectstest.B12Push(g, me.ID, "Billowing Shriekmass", "Creature — Spirit", fdnABShriekmassOracle, 2, 3)
	if p, tough := effectstest.EffectivePTOf(t, g, held); p != 2 || tough != 3 {
		t.Errorf("with four cards in the graveyard the Shriekmass is %d/%d, want 2/3", p, tough)
	}
	shriek := effectstest.CastHoldCreature(t, g, "Billowing Shriekmass", "Creature — Spirit", fdnABShriekmassOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.Graveyard.Size(); n != 7 {
		t.Fatalf("graveyard %d, want 7 after milling three", n)
	}
	for _, id := range []uuid.UUID{held, shriek} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 4 {
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
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			for _, p := range g.Seats[1:] {
				p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
			}
			effectstest.AdvanceToMain(t, g)
			hands := make(map[uuid.UUID]int, len(g.Seats))
			for _, p := range g.Seats {
				hands[p.ID] = effectstest.HandSize(p)
			}
			if lost {
				g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
			}
			effectstest.CastAndResolveCreature(t, g, "Bloodtithe Collector", "Creature — Vampire Noble", fdnABCollectorOracle)
			effectstest.PassPriorityAroundTable(t, g)
			for _, p := range g.Seats[1:] {
				if got := effectstest.DiscardChoiceFor(g, p.ID) != nil; got != lost {
					t.Errorf("seat %s asked to discard: %v, want %v", p.ID, got, lost)
				}
				if lost {
					effectstest.DiscardFromHand(t, g, p.ID)
					if effectstest.HandSize(p) != hands[p.ID]-1 {
						t.Errorf("seat %s hand %d -> %d, want one card discarded", p.ID, hands[p.ID], effectstest.HandSize(p))
					}
				}
			}
		})
	}
}

// --- Brass's Bounty -----------------------------------------------------------

func TestFdnABBrasssBountyMakesATreasurePerLandYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		effectstest.PushLand(g, me.ID, "Island")
	}
	effectstest.PushLand(g, opp.ID, "Their Island")
	effectstest.CastCatalogSpell(t, g, "Brass's Bounty", "Sorcery", fdnABBountyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 3 {
		t.Errorf("%d Treasures, want 3", n)
	}
}
