package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s38_cards_test.go — Karlov Manor Commander's disguise, cloak and
// face-up payoffs (#5368, batch 38). Each test plays the real catalog entry.

const (
	s38BoltbenderOracle   = "055e1799-8188-48f4-854e-d3e28363084f"
	s38ExperimentOracle   = "67a10da3-76f2-4f4a-90d1-e77abb5842fe"
	s38PrintlifterOracle  = "3cbed112-9f2b-46b4-a9c3-bfd8a6856f79"
	s38TrueIdentityOracle = "901d3027-4394-481e-9a0e-dec4513d9f88"
	s38KaustOracle        = "af78c13f-67f8-42bf-b012-482a48af4f31"
	s38ShowstopperOracle  = "e16cbcbd-b03a-404c-8d1e-0a05848181d9"
	s38AbsenceOracle      = "71faaee5-70ff-449c-85c8-560eb189a0a7"
	s38AscensionOracle    = "cbf0190a-198b-49cc-8c79-4028da07fa77"
)

func TestCmdr2S38CardsAreRegisteredFull(t *testing.T) {
	disguised := map[string]bool{s38BoltbenderOracle: true, s38ExperimentOracle: true, s38PrintlifterOracle: true, s38TrueIdentityOracle: true}
	for _, o := range []string{s38BoltbenderOracle, s38ExperimentOracle, s38PrintlifterOracle, s38TrueIdentityOracle,
		s38KaustOracle, s38ShowstopperOracle, s38AbsenceOracle, s38AscensionOracle} {
		def, ok := Lookup(o)
		if !ok {
			t.Errorf("%s is not registered", o)
			continue
		}
		if def.Completeness != CompletenessFull {
			t.Errorf("%s completeness = %v, want Full", def.Name, def.Completeness)
		}
		if disguised[o] {
			if alt := game.FaceDownCastFor(o); alt == nil || alt.FaceDown == nil || alt.FaceDown.Kind != game.FaceDownDisguised {
				t.Errorf("%s declares no disguise", def.Name)
			}
		}
	}
}

func TestCmdr2S38ExperimentTwelveDoublesTheCreatureTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmB1CastFaceDown(t, g, "Experiment Twelve", "Creature — Elf Lizard Warrior", s38ExperimentOracle, 4, 4)
	mkmB1TurnFaceUp(t, g, me, id, "G", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4 (its power)", n)
	}
	if p, tough := mkmB1PT(t, g, id); p != 8 || tough != 8 {
		t.Errorf("turned up = %d/%d, want 8/8", p, tough)
	}
}

func TestCmdr2S38ExperimentTwelveGrowsAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Experiment Twelve", "Creature — Elf Lizard Warrior", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Experiment Twelve" {
				g.Battlefield.Cards[i].OracleID = s38ExperimentOracle
			}
		}
	})
	id := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	mkmB1TurnFaceUp(t, g, me, id, "U", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	// Smuggler arrives 6/5 with its own four counters; Twelve adds six more.
	if n := tmtB1Card(t, g, id).Counters["+1/+1"]; n != 4+6 {
		t.Errorf("+1/+1 counters = %d, want 10", n)
	}
}

func TestCmdr2S38PrintlifterOozeMakesAnOozeWithACounterPerOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", "Creature — Bear", 2, 2)
	id := mkmB1CastFaceDown(t, g, "Printlifter Ooze", "Creature — Ooze", s38PrintlifterOracle, 2, 2)
	mkmB1TurnFaceUp(t, g, me, id, "G", "C", "C", "C")
	passPriorityAroundTable(t, g)
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ooze" && c.IsToken() && c.Controller == me.ID {
			found = true
			if n := c.Counters["+1/+1"]; n != 3 {
				t.Errorf("Ooze counters = %d, want 3 (two Bears and the Printlifter)", n)
			}
			if !game.HasKeyword(&c, "trample") {
				t.Error("the Ooze token lacks trample")
			}
		}
	}
	if !found {
		t.Fatal("no Ooze token was created")
	}
}

func TestCmdr2S38TrueIdentityScriesAndDrawsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 5; i++ {
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	id := mkmB1CastFaceDown(t, g, "True Identity", "Enchantment", s38TrueIdentityOracle, 0, 0)
	b := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	before := me.Hand.Size()
	mkmB1TurnFaceUp(t, g, me, id, "W")
	passPriorityAroundTable(t, g)
	answerScryIfAsked(t, g, me)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand = %d, want %d after the first face-up", got, before+1)
	}
	mkmB1TurnFaceUp(t, g, me, b, "U", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	answerScryIfAsked(t, g, me)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand = %d, want %d: the ability triggers once each turn", got, before+1)
	}
}

func TestCmdr2S38BoltbenderRedirectsAnotherSpellWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[g.Turn.ActiveSeat], g.Seats[1].ID, g.Seats[2].ID
	id := mkmB1CastFaceDown(t, g, "Boltbender", "Creature — Goblin Wizard", s38BoltbenderOracle, 4, 2)
	boltFrom(t, g, me, victimA)
	fillPoolColored(me, "R", 1)
	mkmB1TurnFaceUp(t, g, me, id, "R", "C")
	passPriorityAroundTable(t, g)
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Boltbender's trigger opened no retarget prompt")
	}
	if prompt.PickTargetMin != 0 {
		t.Errorf("min = %d, want 0: any number may be left alone", prompt.PickTargetMin)
	}
	answerRetargetWith(t, g, me.ID, rtPlayerRef(victimB))
	passPriorityAroundTable(t, g)
	if lifeOf(g, victimA) != 40 || lifeOf(g, victimB) != 37 {
		t.Errorf("Bolt did not move: A=%d B=%d", lifeOf(g, victimA), lifeOf(g, victimB))
	}
}

func TestCmdr2S38BoltbenderMayDeclineEveryOffer(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA := g.Seats[g.Turn.ActiveSeat], g.Seats[1].ID
	id := mkmB1CastFaceDown(t, g, "Boltbender", "Creature — Goblin Wizard", s38BoltbenderOracle, 4, 2)
	boltFrom(t, g, me, victimA)
	mkmB1TurnFaceUp(t, g, me, id, "R", "C")
	passPriorityAroundTable(t, g)
	answerRetargetWith(t, g, me.ID, nil)
	passPriorityAroundTable(t, g)
	if lifeOf(g, victimA) != 37 {
		t.Errorf("declined: A=%d, want 37", lifeOf(g, victimA))
	}
}

func TestCmdr2S38ShowstoppingSurpriseTurnsUpThenDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, me.ID, "Bystander", "Creature — Bear", 5, 5)
	floatForTest(g, me, "CCCRR")
	castCatalogSpell(t, g, "Showstopping Surprise", "Instant", s38ShowstopperOracle, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	up := tmtB1Card(t, g, mine)
	if up.FaceDown {
		t.Error("the target is still face down")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("their Bear survived 6 damage")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("the 5/5 died to 6 damage? it should (6>=5)")
	}
}

func TestCmdr2S38UnexplainedAbsenceExilesPerPlayerAndTheyCloak(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Mine Top", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Their Top", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	floatForTest(g, me, "CCCW")
	castCatalogSpell(t, g, "Unexplained Absence", "Instant", s38AbsenceOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("a target was not exiled")
	}
	cloaked := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.FaceDownIsPermanent() {
			cloaked[c.Controller]++
		}
	}
	if cloaked[me.ID] != 1 || cloaked[opp.ID] != 1 {
		t.Errorf("cloaked per player = %v, want one each", cloaked)
	}
}

func TestCmdr2S38VeiledAscensionGivesFaceDownCreaturesFlyingCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	early := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	castHoldCreature(t, g, "Veiled Ascension", "Enchantment", s38AscensionOracle, 0, 0)
	passPriorityAroundTable(t, g)
	if n := tmtB1Card(t, g, early).Counters[game.CounterFlying]; n != 1 {
		t.Errorf("the face-down creature on the battlefield has %d flying counters, want 1", n)
	}
	late := mkmB1CastFaceDown(t, g, "Bubble Smuggler", "Creature — Octopus Fish", mkmBubbleSmugglerOracle, 2, 1)
	if n := tmtB1Card(t, g, late).Counters[game.CounterFlying]; n != 1 {
		t.Errorf("a face-down creature entering later has %d flying counters, want 1", n)
	}
	_ = me
}
