package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b4_test.go — Planechase plane cards, batch 4 (ADR 0216 slice 4,
// #4870). Every test plays the real catalog entry face up on a Planechase
// table (planes_harness_test.go): the plane's static and triggered abilities
// from the command zone, and its chaos ability on a chaos roll.

var (
	planeGrove         = planeFixture{"3e553881-bac7-4830-b739-3b8c6aa0ded6", "Grove of the Dreampods", "Plane — Fabacin"}
	planeHedronFields  = planeFixture{"1f7daaf2-3206-4f52-928a-f5116b782367", "Hedron Fields of Agadeem", "Plane — Zendikar"}
	planeInysHaen      = planeFixture{"61c5f2cb-f39c-487c-9786-a8dad4663f4f", "Inys Haen", "Plane — Cridhe"}
	planeIsleOfVesuva  = planeFixture{"f017380f-97db-4a36-b111-b257aa17ebaf", "Isle of Vesuva", "Plane — Dominaria"}
	planeJund          = planeFixture{"9aa94de8-b114-4479-a771-c7bc0297be58", "Jund", "Plane — Alara"}
	planeKetria        = planeFixture{"f3b27908-ca10-4793-ab8e-6639c73cfadd", "Ketria", "Plane — Ikoria"}
	planeKharasha      = planeFixture{"256770fe-c17d-45dd-bf05-36eeee240324", "Kharasha Foothills", "Plane — Mongseng"}
	planeKrosa         = planeFixture{"2732b605-7c88-4995-b46e-62dc09e68919", "Krosa", "Plane — Dominaria"}
	planeLittjara      = planeFixture{"0ff2bbed-e435-4108-b02a-564a75abba40", "Littjara", "Plane — Kaldheim"}
	planeMegafloraJung = planeFixture{"c907333b-eae8-444f-9da3-b193289f19e2", "Megaflora Jungle", "Plane — Gargantikar"}
	planeNaktamun      = planeFixture{"a2048704-0d22-435d-b819-269291f2981e", "Naktamun", "Plane — Amonkhet"}
	planeNewArgive     = planeFixture{"08427dab-3874-4a29-bd63-cbd16c1d229d", "New Argive", "Plane — Dominaria"}
	planeNornsSeedcore = planeFixture{"ab72ff80-738a-4468-aecf-5d806143f791", "Norn's Seedcore", "Plane — New Phyrexia"}
	planeNyx           = planeFixture{"6552e3dd-ce35-4267-a25e-f5d4dceb8311", "Nyx", "Plane — Theros"}
)

func planeLibrary(me *game.Player, cards ...game.Card) []uuid.UUID { return setLibraryTopFirst(me, cards...) }

// --- Krosa -------------------------------------------------------------------

func TestKrosaPumpsEveryCreatureAndChaosMayAddFiveColors(t *testing.T) {
	requirePlaneRegisteredFull(t, planeKrosa)
	g, _ := newPlaneGame(t, planeKrosa)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	for id, want := range map[uuid.UUID]int{mine: 4, theirs: 3} {
		if c := tmtB1Card(t, g, id); c.CurrentPower() != want || c.CurrentToughness() != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want, want)
		}
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("the pool holds %d mana after the chaos ability, want 5", got)
	}
	// Declining adds nothing.
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("declining changed the pool to %d", got)
	}
}

// --- Megaflora Jungle ----------------------------------------------------------

func TestMegafloraJunglePumpsSmallCreaturesAndChaosMakesAButterfly(t *testing.T) {
	requirePlaneRegisteredFull(t, planeMegafloraJung)
	g, _ := newPlaneGame(t, planeMegafloraJung)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	small := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if c := tmtB1Card(t, g, small); c.CurrentPower() != 4 {
		t.Errorf("the mana value 2 creature is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if c := tmtB1Card(t, g, big); c.CurrentPower() != 2 {
		t.Errorf("the mana value 3 creature is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	var bfly game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Butterfly" {
			bfly = c
		}
	}
	if bfly.InstanceID == uuid.Nil || !bfly.HasSubtypeWord("Insect") || bfly.Controller != me.ID {
		t.Fatalf("no Butterfly Insect under the planar controller: %+v", bfly)
	}
	if !hasEffectiveKeyword(t, g, bfly.InstanceID, "flying") {
		t.Error("the Butterfly lacks flying")
	}
	// A token has mana value 0: the Butterfly is 3/3 under the plane.
	if c := tmtB1Card(t, g, bfly.InstanceID); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("the Butterfly is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Hedron Fields of Agadeem ----------------------------------------------------

func TestHedronFieldsStopsBigCreaturesAndChaosMakesAnEldrazi(t *testing.T) {
	requirePlaneRegisteredFull(t, planeHedronFields)
	g, _ := newPlaneGame(t, planeHedronFields)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	huge := pushVanillaCreature(g, me.ID, "Huge", 7, 7)
	almost := pushVanillaCreature(g, me.ID, "Almost", 6, 6)
	pump := pushVanillaCreature(g, me.ID, "Pumpable", 5, 5)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 8, 8)
	g.BumpLayerVersionForTest()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(huge, opp.ID); err == nil {
		t.Error("a 7/7 was allowed to attack")
	}
	if err := g.DeclareAttacker(almost, opp.ID); err != nil {
		t.Errorf("a 6/6 was refused: %v", err)
	}
	// A pump to power 7 takes a creature out of combat before attackers
	// are declared, and a -X/-X puts it back.
	if err := (BoostUntilEOT{Target: pump, Power: 2, Toughness: 0}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
		t.Fatalf("pump: %v", err)
	}
	if err := g.DeclareAttacker(pump, opp.ID); err == nil {
		t.Error("a creature pumped to power 7 was allowed to attack")
	}
	_ = blocker
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	var tok game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi" {
			tok = c
		}
	}
	if tok.InstanceID == uuid.Nil || tok.CurrentPower() != 7 || tok.CurrentToughness() != 7 || len(tok.Colors) != 0 {
		t.Fatalf("no colorless 7/7 Eldrazi token: %+v", tok)
	}
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "annihilator 1") {
		t.Error("the Eldrazi token lacks annihilator 1")
	}
}

// --- Panopticon-like: Grove of the Dreampods -----------------------------------------

func TestGroveOfTheDreampodsRevealsToACreatureAndChaosReanimates(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGrove)
	g, _ := newPlaneGame(t, planeGrove)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := planeLibrary(me, libraryLand("Forest"), libraryLand("Forest"), libraryCreature("Hit"), libraryLand("Deep"))
	planeswalkToPlane(t, g, planeGrove.Oracle)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[2]) {
		t.Fatal("the creature card revealed did not enter the battlefield")
	}
	if me.Library.Size() != 3 {
		t.Errorf("the library has %d cards, want 3", me.Library.Size())
	}
	for _, id := range ids[:2] {
		if !me.Library.Contains(id) {
			t.Errorf("a revealed land left the library")
		}
	}
	// No creature in the library: everything goes to the bottom, nothing enters.
	planeLibrary(me, libraryLand("A"), libraryLand("B"))
	before := len(g.Battlefield.Cards)
	_ = toNextUpkeep(t, g)
	next := g.Seats[g.Turn.ActiveSeat]
	planeLibrary(next, libraryLand("A"), libraryLand("B"))
	passPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before {
		t.Errorf("a land entered the battlefield")
	}

	// Chaos returns a creature card from the graveyard to the battlefield.
	g2, _ := newPlaneGame(t, planeGrove)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	bear := pushGraveyardCardWithTypeLineID(me2, "Bear", "Creature — Bear")
	rock := pushGraveyardCardWithTypeLineID(me2, "Rock", "Artifact")
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	ch := pickTargetPrompt(t, g2)
	if err := g2.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(rock)); err == nil {
		t.Fatal("a noncreature card was accepted")
	}
	answerPickTarget(t, g2, bear)
	passPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(bear) {
		t.Error("the creature card did not return to the battlefield")
	}
}

// --- Inys Haen -----------------------------------------------------------------------

func TestInysHaenMillsThenReturnsEveryPlayersLandsWhenYouLeave(t *testing.T) {
	requirePlaneRegisteredFull(t, planeInysHaen)
	g, _ := newPlaneGame(t, planeInysHaen)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	size := me.Library.Size()
	planeswalkToPlane(t, g, planeInysHaen.Oracle)
	// The harness walked away from Inys Haen to put it back on top: its own
	// planeswalk-away ability is waiting beside the planeswalk-to one.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := size - me.Library.Size(); got != 3 {
		t.Fatalf("%d cards milled on the planeswalk, want 3", got)
	}
	myLand := pushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
	hisLand := pushGraveyardCardWithTypeLineID(opp, "Island", "Basic Land — Island")
	spell := pushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myLand, hisLand} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a land did not return to the battlefield")
			continue
		}
		if !tmtB1Card(t, g, id).Tapped {
			t.Errorf("a land returned untapped")
		}
	}
	if g.Battlefield.Contains(spell) {
		t.Error("a nonland card returned")
	}
	if c := tmtB1Card(t, g, hisLand); c.Controller != opp.ID {
		t.Error("an opponent's land came back under someone else's control")
	}
}

func TestInysHaenChaosReturnsANonlandCard(t *testing.T) {
	g, _ := newPlaneGame(t, planeInysHaen)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
	bolt := pushGraveyardCardWithTypeLineID(me, "Bolt", "Instant")
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(land)); err == nil {
		t.Fatal("a land card was accepted")
	}
	answerPickTarget(t, g, bolt)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the nonland card did not return to hand")
	}
}

// --- Littjara ---------------------------------------------------------------------------

func TestLittjaraMakesAChangelingAndChaosCountersTheChosenType(t *testing.T) {
	requirePlaneRegisteredFull(t, planeLittjara)
	g, _ := newPlaneGame(t, planeLittjara)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	planeswalkToPlane(t, g, planeLittjara.Oracle)
	passPriorityAroundTable(t, g)
	var shape game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Shapeshifter" && c.Controller == me.ID {
			shape = c
		}
	}
	if shape.InstanceID == uuid.Nil || !hasEffectiveKeyword(t, g, shape.InstanceID, "changeling") {
		t.Fatal("no changeling Shapeshifter")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	goblin := b12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Bear")
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{bear: 1, shape.InstanceID: 1, goblin: 0, theirs: 0} {
		if got := tmtB1Card(t, g, id).Counters[game.CounterPlusOne]; got != want {
			t.Errorf("%s has %d +1/+1 counters, want %d", tmtB1Card(t, g, id).Name, got, want)
		}
	}
}

// --- New Argive ----------------------------------------------------------------------------

func TestNewArgivePumpsHistoricAttackersAndChaosFindsAHistoricCard(t *testing.T) {
	requirePlaneRegisteredFull(t, planeNewArgive)
	g, _ := newPlaneGame(t, planeNewArgive)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	g.BumpLayerVersionForTest()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{legend, plain} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, legend); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the historic attacker is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if c := tmtB1Card(t, g, plain); c.CurrentPower() != 2 {
		t.Errorf("a nonhistoric attacker is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}

	g2, _ := newPlaneGame(t, planeNewArgive)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	ids := planeLibrary(me2, libraryLand("A"), libraryCreature("B"),
		game.Card{Name: "Relic", TypeLine: "Artifact"}, libraryLand("C"))
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	if !me2.Hand.Contains(ids[2]) {
		t.Fatal("the historic card did not go to hand")
	}
	if me2.Library.Size() != 3 {
		t.Errorf("the library holds %d cards, want 3 (two revealed to the bottom and the card beneath)", me2.Library.Size())
	}
	if top := me2.Library.Cards[len(me2.Library.Cards)-1]; top.InstanceID != ids[3] {
		t.Errorf("the card below the revealed ones is not on top of the library")
	}
}
