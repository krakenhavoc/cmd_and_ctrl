package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b19_cards_test.go — Standard leftovers batches 19 and 20 (#4465,
// Foundations): one engine-driven test per card clause.

const (
	highbornVampireOracle    = "1adcf587-9035-4299-a2bd-62f805fdd897"
	undyingMaliceOracle      = "8b061702-1bc3-45aa-a758-446b25034801"
	dropkickBomberOracle     = "c394d4b6-cfb8-4639-90cb-a322a25c0a2f"
	ragingRedcapOracle       = "679841ad-f3ac-48e2-ba11-2cf6d78642b4"
	bearCubOracle            = "ed206651-3d9b-4546-8d45-0682817192fd"
	knightOfGraceOracle      = "7b51a979-47e0-4b98-8ec5-8064c83422e1"
	releaseTheDogsOracle     = "b97d21b4-d730-479f-872f-3e1645b66751"
	kalastriaHighbornOracle  = "77e319da-d782-4258-bf0e-626755e5adc9"
	knightOfMaliceOracle     = "b89d912a-65a0-4bc9-9e1d-ff8317458a72"
	myojinNightsReachOracle  = "5ac65688-0a74-4dc4-b6cb-12bdf8aed01e"
	tributeToHungerOracle    = "dc5a501c-e16c-4d4b-a510-70a14272473e"
	harmlessOfferingOracle   = "f5d32e5f-066b-42cf-8dd1-072a96275313"
	hoardingDragonOracle     = "d36d3c93-ed24-4447-9d15-3071f5c0989b"
	ravenousGiantOracle      = "46c7a552-4a98-4e1a-8651-8fb0e7153aec"
	ordealOfNyleaOracle      = "5fff2bf4-9e26-4815-9bf6-672ed7e55b96"
	venomConnoisseurOracle   = "03ebc100-68ac-4d53-adc2-d537212f7d31"
	wildbornPreserverOracle  = "55cec4bc-3de6-4c0d-a4cc-c5a30849fbda"
	dryadMilitantOracle      = "b8ca5877-ac9e-4b15-8c23-c70f61b01895"
	ovikaEnigmaGoliathOracle = "b05a2b5f-1b5c-4f1b-8a6b-525c5c8b8ba1"
	primeSpeakerZeganaOracle = "311e9368-696a-47e7-aa2f-3ef1b1a92e2b"
)

func TestStdB19CardsAreCompleteWithNoCaveats(t *testing.T) {
	for name, oracle := range map[string]string{
		"Highborn Vampire": highbornVampireOracle, "Undying Malice": undyingMaliceOracle,
		"Dropkick Bomber": dropkickBomberOracle, "Raging Redcap": ragingRedcapOracle,
		"Bear Cub": bearCubOracle, "Knight of Grace": knightOfGraceOracle,
		"Release the Dogs": releaseTheDogsOracle, "Kalastria Highborn": kalastriaHighbornOracle,
		"Knight of Malice": knightOfMaliceOracle, "Myojin of Night's Reach": myojinNightsReachOracle,
		"Tribute to Hunger": tributeToHungerOracle, "Harmless Offering": harmlessOfferingOracle,
		"Hoarding Dragon": hoardingDragonOracle, "Ravenous Giant": ravenousGiantOracle,
		"Ordeal of Nylea": ordealOfNyleaOracle, "Venom Connoisseur": venomConnoisseurOracle,
		"Wildborn Preserver": wildbornPreserverOracle, "Dryad Militant": dryadMilitantOracle,
		"Ovika, Enigma Goliath": ovikaEnigmaGoliathOracle, "Prime Speaker Zegana": primeSpeakerZeganaOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want CompletenessFull with no caveats, got %v %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestStdB19RagingRedcapHasDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	id := pushCatalogPermanent(g, g.Seats[0].ID, "Raging Redcap", "Creature — Goblin Knight", ragingRedcapOracle, false)
	if !effectiveAbilitiesContain(t, g, id, "double strike") {
		t.Errorf("abilities %v lack double strike", effectiveAbilities(t, g, id))
	}
}

// Knight of Grace: +1/+0 while ANY player controls a black permanent, one
// black permanent or three the same; hexproof from black refuses a black
// opposing source and nothing else.
func TestStdB19KnightOfGraceIsPumpedByAnyBlackPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushTestPermanent(g, me.ID, game.Card{Name: "Knight of Grace", OracleID: knightOfGraceOracle,
		TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2, Colors: []string{"W"}})
	if p, _ := effectivePTOf(t, g, knight); p != 2 {
		t.Fatalf("power %d with no black permanent, want 2", p)
	}
	pushTestPermanent(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	if p, _ := effectivePTOf(t, g, knight); p != 2 {
		t.Errorf("power %d with only a green permanent, want 2", p)
	}
	pushTestPermanent(g, opp.ID, game.Card{Name: "Black A", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"B"}})
	if p, tough := effectivePTOf(t, g, knight); p != 3 || tough != 2 {
		t.Errorf("%d/%d with an opponent's black permanent, want 3/2", p, tough)
	}
	pushTestPermanent(g, me.ID, game.Card{Name: "Black B", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"B"}})
	if p, _ := effectivePTOf(t, g, knight); p != 3 {
		t.Errorf("power %d with two black permanents, want 3 (the bonus does not stack)", p)
	}
	if !effectiveAbilitiesContain(t, g, knight, "first strike") {
		t.Errorf("abilities %v lack first strike", effectiveAbilities(t, g, knight))
	}
	if !effectiveAbilitiesContain(t, g, knight, game.HexproofFromColor("B")) {
		t.Errorf("abilities %v lack hexproof from black", effectiveAbilities(t, g, knight))
	}
	if creatureTargetableFrom(g, opp.ID, knight, srcBlack) {
		t.Error("an opponent's black source can target Knight of Grace")
	}
	if !creatureTargetableFrom(g, opp.ID, knight, srcRed) {
		t.Error("an opponent's red source can't target Knight of Grace")
	}
	if !creatureTargetableFrom(g, me.ID, knight, srcBlack) {
		t.Error("its own controller's black source can't target it: hexproof from only stops opponents")
	}
}

func TestStdB19KnightOfMaliceMirrorsGraceWithWhite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := pushTestPermanent(g, me.ID, game.Card{Name: "Knight of Malice", OracleID: knightOfMaliceOracle,
		TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2, Colors: []string{"B"}})
	if p, _ := effectivePTOf(t, g, knight); p != 2 {
		t.Fatalf("power %d with no white permanent, want 2 (it is black itself)", p)
	}
	pushTestPermanent(g, opp.ID, game.Card{Name: "White Angel", TypeLine: "Creature — Angel", Power: 1, Toughness: 1, Colors: []string{"W"}})
	if p, _ := effectivePTOf(t, g, knight); p != 3 {
		t.Errorf("power %d with a white permanent, want 3", p)
	}
	if creatureTargetableFrom(g, opp.ID, knight, &game.Characteristic{Colors: []string{"W"}, Types: []string{"Instant"}}) {
		t.Error("an opponent's white source can target Knight of Malice")
	}
	if !creatureTargetableFrom(g, opp.ID, knight, srcBlack) {
		t.Error("an opponent's black source can't target Knight of Malice")
	}
}

func TestStdB19ReleaseTheDogsMakesFourDogs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "WWWW")
	castFromHandForTest(t, g, me, "Release the Dogs", "Sorcery", "{3}{W}", releaseTheDogsOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Dog", me.ID); n != 4 {
		t.Errorf("%d Dogs, want 4", n)
	}
}

func TestStdB19RavenousGiantDamagesItsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushTestPermanent(g, me.ID, game.Card{Name: "Ravenous Giant", OracleID: ravenousGiantOracle,
		TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	// The first turn's upkeep is behind us: walk to the next one of mine.
	for i := 0; i < 600; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.ActiveSeat == me.Seat && g.Turn.Step == game.StepUpkeep && g.Turn.Round > 1 {
			break
		}
	}
	before, oppBefore := me.Life, opp.Life
	passPriorityAroundTable(t, g)
	if got := before - me.Life; got != 1 {
		t.Errorf("lost %d life at my upkeep, want 1", got)
	}
	if opp.Life != oppBefore {
		t.Errorf("an opponent's life changed: %d -> %d", oppBefore, opp.Life)
	}
}

// --- Kalastria Highborn -----------------------------------------------

func kalastriaTable(t *testing.T) (g *game.Game, me, opp *game.Player, highborn uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	highborn = pushTestPermanent(g, me.ID, game.Card{Name: "Kalastria Highborn", OracleID: kalastriaHighbornOracle,
		TypeLine: "Creature — Vampire Shaman", Power: 2, Toughness: 2, Colors: []string{"B"}})
	return g, me, opp, highborn
}

// settleKalastria targets opp with the dies trigger and answers the {B}.
func settleKalastria(t *testing.T, g *game.Game, me, opp *game.Player, pay bool) {
	t.Helper()
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	assistCardSettle(t, g, me.ID, &pick, true)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no {B} payment was offered")
	}
	answerPayUnless(t, g, me.ID, pay)
	passPriorityAroundTable(t, g)
}

func TestStdB19KalastriaHighbornDrainsWhenAnotherVampireDiesAndBIsPaid(t *testing.T) {
	g, me, opp, _ := kalastriaTable(t)
	vamp := pushTestPermanent(g, me.ID, game.Card{Name: "Fang", TypeLine: "Creature — Vampire", Power: 2, Toughness: 2, Colors: []string{"B"}})
	floatForTest(g, me, "B")
	myLife, oppLife := me.Life, opp.Life
	destroyNow(t, g, vamp)
	settleKalastria(t, g, me, opp, true)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("life %d/%d, want %d/%d (target loses 2, I gain 2)", me.Life, opp.Life, myLife+2, oppLife-2)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, the {B} was not paid", got)
	}
}

func TestStdB19KalastriaHighbornDecliningThePaymentDoesNothing(t *testing.T) {
	g, me, opp, _ := kalastriaTable(t)
	vamp := pushTestPermanent(g, me.ID, game.Card{Name: "Fang", TypeLine: "Creature — Vampire", Power: 2, Toughness: 2, Colors: []string{"B"}})
	floatForTest(g, me, "B")
	myLife, oppLife := me.Life, opp.Life
	destroyNow(t, g, vamp)
	settleKalastria(t, g, me, opp, false)
	if opp.Life != oppLife || me.Life != myLife {
		t.Errorf("life changed to %d/%d without paying", me.Life, opp.Life)
	}
}

func TestStdB19KalastriaHighbornTriggersOnItsOwnDeath(t *testing.T) {
	g, me, opp, highborn := kalastriaTable(t)
	floatForTest(g, me, "B")
	oppLife := opp.Life
	destroyNow(t, g, highborn)
	settleKalastria(t, g, me, opp, true)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, oppLife-2)
	}
}

func TestStdB19KalastriaHighbornIgnoresNonVampiresAndOpposingVampires(t *testing.T) {
	g, me, opp, _ := kalastriaTable(t)
	bear := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Fang", TypeLine: "Creature — Vampire", Power: 2, Toughness: 2})
	floatForTest(g, me, "B")
	destroyNow(t, g, bear)
	destroyNow(t, g, theirs)
	passPriorityAroundTable(t, g)
	if len(g.StackMeta) != 0 || hasPayUnlessFor(g, me.ID) || latestPickTarget(g, me.ID) != nil {
		t.Error("a Bear or an opposing Vampire dying triggered Kalastria Highborn")
	}
}

func TestStdB19KalastriaHighbornSeesAChangelingAsAVampire(t *testing.T) {
	g, me, opp, _ := kalastriaTable(t)
	ling := pushTestPermanent(g, me.ID, game.Card{Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter", Power: 2, Toughness: 2, Keywords: []string{"changeling"}})
	floatForTest(g, me, "B")
	oppLife := opp.Life
	destroyNow(t, g, ling)
	settleKalastria(t, g, me, opp, true)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, oppLife-2)
	}
}

// --- Undying Malice ---------------------------------------------------

func TestStdB19UndyingMaliceReturnsTheCreatureTappedWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	dgCastOn(t, g, "Undying Malice", undyingMaliceOracle, bear)
	if grantedTriggerCount(g, bear) != 1 {
		t.Fatal("the target did not gain the dies trigger")
	}
	b18Kill(t, g, bear)
	back := findBattlefieldCardByID(g, bear)
	if back == nil {
		t.Fatal("the creature died and did not come back")
	}
	if !back.Tapped {
		t.Error("it returns tapped")
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	b18Kill(t, g, bear)
	if findBattlefieldCardByID(g, bear) != nil {
		t.Error("the second death is final: the returned creature is a new object")
	}
}

// --- Dropkick Bomber --------------------------------------------------

func bomberTable(t *testing.T) (g *game.Game, me, opp *game.Player, bomber, goblin uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bomber = pushTestPermanent(g, me.ID, game.Card{Name: "Dropkick Bomber", OracleID: dropkickBomberOracle,
		TypeLine: "Creature — Goblin Warrior", Power: 2, Toughness: 3, Colors: []string{"R"}})
	goblin = pushTestPermanent(g, me.ID, game.Card{Name: "Goblin Piker", TypeLine: "Creature — Goblin", Power: 2, Toughness: 1, Colors: []string{"R"}})
	return g, me, opp, bomber, goblin
}

func TestStdB19DropkickBomberPumpsOtherGoblinsYouControlOnly(t *testing.T) {
	g, me, opp, bomber, goblin := bomberTable(t)
	bear := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 1})
	for name, tc := range map[string]struct {
		id         uuid.UUID
		power, tgh int
	}{
		"the Bomber itself":  {bomber, 2, 3},
		"another Goblin":     {goblin, 3, 2},
		"a non-Goblin":       {bear, 2, 2},
		"an opposing Goblin": {theirs, 2, 1},
	} {
		if p, tough := effectivePTOf(t, g, tc.id); p != tc.power || tough != tc.tgh {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, tc.power, tc.tgh)
		}
	}
}

func TestStdB19DropkickBomberGrantsFlyingAndASacrificeOnCombatDamage(t *testing.T) {
	g, me, opp, bomber, goblin := bomberTable(t)
	floatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, bomber, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: goblin}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, goblin, "flying") {
		t.Fatalf("the Goblin did not gain flying: %v", effectiveAbilities(t, g, goblin))
	}
	if grantedTriggerCount(g, goblin) != 1 {
		t.Fatal("the Goblin did not gain the sacrifice trigger")
	}
	attackWith(t, g, opp.ID, goblin)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, goblin) != nil {
		t.Error("the Goblin was not sacrificed after dealing combat damage")
	}
	if findBattlefieldCardByID(g, bomber) == nil {
		t.Error("the Bomber is gone")
	}
}

func TestStdB19DropkickBomberCantTargetItselfOrANonGoblin(t *testing.T) {
	g, me, _, bomber, _ := bomberTable(t)
	bear := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	floatForTest(g, me, "RR")
	for name, id := range map[string]uuid.UUID{"itself": bomber, "a Bear": bear} {
		err := g.ActivateCatalogAbility(me.ID, bomber, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}})
		if err == nil {
			t.Errorf("targeting %s was accepted", name)
		}
	}
}

// --- Myojin of Night's Reach ------------------------------------------

func TestStdB19MyojinEntersWithADivinityCounterOnlyWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "BBBBBBBB")
	id := castFromHand(t, g, game.Card{Name: "Myojin of Night's Reach", TypeLine: "Legendary Creature — Spirit", ManaCost: "{5}{B}{B}{B}", OracleID: myojinNightsReachOracle, Power: 5, Toughness: 2, Colors: []string{"B"}}, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, "divinity"); got != 1 {
		t.Fatalf("divinity counters %d after being cast from hand, want 1", got)
	}
	if !effectiveAbilitiesContain(t, g, id, "indestructible") {
		t.Errorf("no indestructible with a divinity counter: %v", effectiveAbilities(t, g, id))
	}

	// Put onto the battlefield without being cast: no counter, not indestructible.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	plain := hobEnter(t, g2, me2.ID, game.Card{Name: "Myojin of Night's Reach", OracleID: myojinNightsReachOracle,
		TypeLine: "Legendary Creature — Spirit", Power: 5, Toughness: 2})
	if got := counterCount(g2, plain, "divinity"); got != 0 {
		t.Errorf("divinity counters %d without being cast, want 0", got)
	}
	if effectiveAbilitiesContain(t, g2, plain, "indestructible") {
		t.Error("indestructible without a divinity counter")
	}
}

func TestStdB19MyojinRemovingTheCounterMakesEachOpponentDiscardTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "BBBBBBBB")
	id := castFromHand(t, g, game.Card{Name: "Myojin of Night's Reach", TypeLine: "Legendary Creature — Spirit", ManaCost: "{5}{B}{B}{B}", OracleID: myojinNightsReachOracle, Power: 5, Toughness: 2, Colors: []string{"B"}}, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	myHand := me.Hand.Size()
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Hand.Size() == 0 {
			t.Fatalf("seat %d starts with an empty hand", p.Seat)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterCount(g, id, "divinity"); got != 0 {
		t.Errorf("the counter is the cost and is gone: %d left", got)
	}
	if effectiveAbilitiesContain(t, g, id, "indestructible") {
		t.Error("still indestructible with no divinity counter")
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			if p.Hand.Size() != myHand {
				t.Errorf("my hand went from %d to %d", myHand, p.Hand.Size())
			}
			continue
		}
		if p.Hand.Size() != 0 {
			t.Errorf("seat %d still holds %d cards", p.Seat, p.Hand.Size())
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation with no counter was accepted")
	}
}

// --- Tribute to Hunger ------------------------------------------------

func TestStdB19TributeToHungerEdictGainsLifeEqualToToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	big := pushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 6})
	advanceToMain(t, g)
	floatForTest(g, me, "BBB")
	life := me.Life
	castFromHandForTest(t, g, me, "Tribute to Hunger", "Instant", "{2}{B}", tributeToHungerOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	// With a single creature the opponent may still be asked to choose it.
	answerSacrifice(t, g, opp.ID, big)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, big) != nil {
		t.Fatal("the creature was not sacrificed")
	}
	if me.Life != life+6 {
		t.Errorf("life %d, want %d (the Wall's toughness)", me.Life, life+6)
	}
}

func TestStdB19TributeToHungerGainsNothingFromAnEmptyBoardAndCantTargetYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	floatForTest(g, me, "BBBBBB")
	life := me.Life
	castFromHandForTest(t, g, me, "Tribute to Hunger", "Instant", "{2}{B}", tributeToHungerOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("gained %d life from a player with no creature", me.Life-life)
	}
	if _, err := castFromHandErr(g, me, "Tribute to Hunger", "Instant", "{2}{B}", tributeToHungerOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("targeting yourself was accepted: it says target opponent")
	}
}

// --- Harmless Offering ------------------------------------------------

func TestStdB19HarmlessOfferingGivesAPermanentToTheOpponentForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	floatForTest(g, me, "RRR")
	castFromHandForTest(t, g, me, "Harmless Offering", "Sorcery", "{2}{R}", harmlessOfferingOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, bear)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("the Bear's controller is %v, want the opponent", c)
	}
	if c.Owner != me.ID {
		t.Error("ownership changed: it is only control")
	}
	// Still theirs on the next turn: no duration.
	for i := 0; i < 12; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if c := findBattlefieldCardByID(g, bear); c == nil || c.Controller != opp.ID {
		t.Error("the Bear came back at end of turn: the gift has no duration")
	}
}

func TestStdB19HarmlessOfferingRefusesAnOpponentsPermanentAndYourself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	floatForTest(g, me, "RRRRRR")
	for name, targets := range map[string][]game.TargetRef{
		"an opposing permanent": {{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: theirs}},
		"yourself":              {{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetCard, ID: mine}},
	} {
		if _, err := castFromHandErr(g, me, "Harmless Offering", "Sorcery", "{2}{R}", harmlessOfferingOracle,
			game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
			t.Errorf("targeting %s was accepted", name)
		}
	}
}

// --- Hoarding Dragon --------------------------------------------------

func hoardingDragonTable(t *testing.T) (g *game.Game, me *game.Player, dragon, sol, forest uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"})
	forest, sol = ids[0], ids[1]
	advanceToMain(t, g)
	floatForTest(g, me, "RRRRR")
	dragon = castFromHand(t, g, game.Card{Name: "Hoarding Dragon", TypeLine: "Creature — Dragon", ManaCost: "{3}{R}{R}",
		OracleID: hoardingDragonOracle, Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying"}},
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	return g, me, dragon, sol, forest
}

func TestStdB19HoardingDragonExilesAnArtifactAndReturnsItToHandWhenItDies(t *testing.T) {
	g, me, dragon, sol, forest := hoardingDragonTable(t)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt after the Dragon entered")
	}
	if got := searchOptionNamed(g, searchChoiceFor(g, me.ID), "Forest"); got != uuid.Nil {
		t.Error("a non-artifact card was offered")
	}
	answerSearchNamed(t, g, me.ID, "Sol Ring")
	passPriorityAroundTable(t, g)
	if !assistCardInZone(g, sol, game.ZoneExile) {
		t.Fatal("the artifact was not exiled")
	}
	if c, ok := g.LookupCardForEffect(sol); !ok || c.FaceDown {
		t.Error("the exiled artifact is face down: the ruling says face up")
	}
	if !assistCardInZone(g, forest, game.ZoneLibrary) {
		t.Error("the Forest left the library")
	}
	destroyNow(t, g, dragon)
	assistCardSettle(t, g, me.ID, nil, true)
	if !me.Hand.Contains(sol) {
		t.Error("the exiled artifact did not return to its owner's hand")
	}
}

func TestStdB19HoardingDragonMayDeclineBothHalves(t *testing.T) {
	g, me, dragon, sol, _ := hoardingDragonTable(t)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !assistCardInZone(g, sol, game.ZoneLibrary) {
		t.Fatal("declining the search still moved the artifact")
	}
	destroyNow(t, g, dragon)
	assistCardSettle(t, g, me.ID, nil, true)
	if me.Hand.Contains(sol) {
		t.Error("a hoard that was never made came to hand")
	}

	// And a hoard that was made stays exiled when the "you may" is declined.
	g2, me2, dragon2, sol2, _ := hoardingDragonTable(t)
	answerSearchNamed(t, g2, me2.ID, "Sol Ring")
	passPriorityAroundTable(t, g2)
	destroyNow(t, g2, dragon2)
	assistCardSettle(t, g2, me2.ID, nil, false)
	if me2.Hand.Contains(sol2) || !assistCardInZone(g2, sol2, game.ZoneExile) {
		t.Error("declining the dies trigger still returned the card")
	}
}

// --- Ordeal of Nylea --------------------------------------------------

func ordealTable(t *testing.T, counters int) (g *game.Game, me, opp *game.Player, creature, aura uuid.UUID, lands []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	lands = seedSearchLibrary(me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	creature = pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if counters > 0 {
		findBattlefieldCardForTest(g, creature).Counters = map[string]int{game.CounterPlusOne: counters}
	}
	aura = castCatalogSpell(t, g, "Ordeal of Nylea", "Enchantment — Aura", ordealOfNyleaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: creature}})
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, aura); c == nil || !c.IsAttachedTo(creature) {
		t.Fatal("the Aura did not attach to the creature")
	}
	return g, me, opp, creature, aura, lands
}

func ordealAttack(t *testing.T, g *game.Game, opp *game.Player, creature uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(creature, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

func TestStdB19OrdealOfNyleaPutsACounterOnEachAttackAndKeepsTheAuraBelowThree(t *testing.T) {
	g, _, opp, creature, aura, _ := ordealTable(t, 0)
	ordealAttack(t, g, opp, creature)
	if got := counterCount(g, creature, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after one attack, want 1", got)
	}
	if findBattlefieldCardByID(g, aura) == nil {
		t.Error("the Aura was sacrificed with one counter")
	}
}

func TestStdB19OrdealOfNyleaSacrificesAtThreeCountersAndFetchesTwoTappedBasics(t *testing.T) {
	g, me, opp, creature, aura, lands := ordealTable(t, 2)
	ordealAttack(t, g, opp, creature)
	if got := counterCount(g, creature, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	if findBattlefieldCardByID(g, aura) != nil {
		t.Fatal("the Aura was not sacrificed at three counters")
	}
	if !inGraveyardOf(g, me.ID, aura) {
		t.Error("the sacrificed Aura is not in the graveyard")
	}
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt for the sacrifice trigger")
	}
	if searchOptionNamed(g, searchChoiceFor(g, me.ID), "Sol Ring") != uuid.Nil {
		t.Error("a non-land card was offered")
	}
	answerSearchByID(t, g, me.ID, lands[0], lands[1])
	passPriorityAroundTable(t, g)
	for _, id := range lands[:2] {
		c := findBattlefieldCardByID(g, id)
		if c == nil {
			t.Fatalf("land %s did not enter the battlefield", id)
		}
		if !c.Tapped {
			t.Errorf("%s entered untapped", c.Name)
		}
	}
	if findBattlefieldCardByID(g, lands[2]) != nil {
		t.Error("a third land came in")
	}
}

// The trigger is "when you sacrifice this Aura" for ANY sacrifice, and the
// state-based fall-off of an Aura whose creature died is not one.
func TestStdB19OrdealOfNyleaSearchesWhenSacrificedForAnyReasonButNotWhenItFallsOff(t *testing.T) {
	g, me, _, _, aura, lands := ordealTable(t, 0)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(aura); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("sacrificing the Aura outside its own ability did not trigger the search")
	}
	answerSearchFailToFind(t, g, me.ID)
	_ = lands

	g2, me2, _, creature2, aura2, _ := ordealTable(t, 0)
	destroyNow(t, g2, creature2)
	g2.RunStateChecksForTest()
	passPriorityAroundTable(t, g2)
	if findBattlefieldCardByID(g2, aura2) != nil {
		t.Fatal("the Aura stayed on the battlefield after its creature died")
	}
	if searchChoiceFor(g2, me2.ID) != nil {
		t.Error("the Aura falling off as a state-based action triggered the search")
	}
}

// --- Venom Connoisseur ------------------------------------------------

func TestStdB19VenomConnoisseurSecondResolutionGivesEveryCreatureDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	venom := pushTestPermanent(g, me.ID, game.Card{Name: "Venom Connoisseur", OracleID: venomConnoisseurOracle,
		TypeLine: "Creature — Human Druid", Power: 2, Toughness: 2, Colors: []string{"G"}})
	bystander := pushTestPermanent(g, me.ID, game.Card{Name: "Bystander", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	enter := func(name string) uuid.UUID {
		id := hobEnter(t, g, me.ID, game.Card{Name: name, TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
		passPriorityAroundTable(t, g)
		return id
	}
	has := func(id uuid.UUID) bool { return effectiveAbilitiesContain(t, g, id, "deathtouch") }

	first := enter("First")
	if !has(venom) || has(bystander) || has(first) {
		t.Errorf("after one creature: venom=%v bystander=%v first=%v, want only the Connoisseur", has(venom), has(bystander), has(first))
	}
	second := enter("Second")
	if !has(venom) || !has(bystander) || !has(first) || !has(second) {
		t.Errorf("after two: venom=%v bystander=%v first=%v second=%v, want all of mine", has(venom), has(bystander), has(first), has(second))
	}
	if has(theirs) {
		t.Error("an opponent's creature gained deathtouch")
	}
	third := enter("Third")
	if has(third) {
		t.Error("the third resolution gave the new creature deathtouch")
	}
}

func TestStdB19VenomConnoisseurIgnoresOpposingAndSelfEntries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	venom := hobEnter(t, g, me.ID, game.Card{Name: "Venom Connoisseur", OracleID: venomConnoisseurOracle,
		TypeLine: "Creature — Human Druid", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	hobEnter(t, g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if effectiveAbilitiesContain(t, g, venom, "deathtouch") {
		t.Error("its own entry or an opponent's creature gave it deathtouch")
	}
}

// --- Wildborn Preserver -----------------------------------------------

func preserverTable(t *testing.T) (g *game.Game, me *game.Player, preserver uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	preserver = pushTestPermanent(g, me.ID, game.Card{Name: "Wildborn Preserver", OracleID: wildbornPreserverOracle,
		TypeLine: "Creature — Elf Archer", Power: 2, Toughness: 2, Colors: []string{"G"}})
	return g, me, preserver
}

func TestStdB19WildbornPreserverPaysXForXCountersWhenANonHumanEnters(t *testing.T) {
	g, me, preserver := preserverTable(t)
	fillPool(me, 3)
	hobEnter(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Errorf("X prompt %d..%d, want 0..3", pa.Min, pa.Max)
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, preserver, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("%d mana left, want 1 (paid {2} of 3)", got)
	}
}

func TestStdB19WildbornPreserverPayingNothingPutsNoCounters(t *testing.T) {
	g, me, preserver := preserverTable(t)
	fillPool(me, 3)
	hobEnter(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, preserver, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters for X=0, want 0", got)
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("%d mana left, want 3", got)
	}
}

func TestStdB19WildbornPreserverIgnoresHumansChangelingsAndOpponents(t *testing.T) {
	g, me, _ := preserverTable(t)
	opp := g.Seats[1]
	fillPool(me, 3)
	hobEnter(t, g, me.ID, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 1, Toughness: 1})
	hobEnter(t, g, me.ID, game.Card{Name: "Hidden Changeling", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1, Keywords: []string{"changeling"}})
	hobEnter(t, g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	passPriorityAroundTable(t, g)
	if c := payAmountPrompt(g, me.ID); c != nil {
		t.Error("a Human, a changeling or an opponent's creature offered the payment")
	}
}

// --- Dryad Militant ---------------------------------------------------

func TestStdB19DryadMilitantExilesInstantAndSorceryCardsInsteadOfTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushTestPermanent(g, opp.ID, game.Card{Name: "Dryad Militant", OracleID: dryadMilitantOracle,
		TypeLine: "Creature — Dryad Soldier", Power: 2, Toughness: 1, Colors: []string{"G", "W"}})
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if !assistCardInZone(g, bolt, game.ZoneExile) {
		t.Errorf("the resolved instant is in %v, want exile", g.FindCardZoneForEffect(bolt).Kind)
	}
	// A discarded sorcery card goes to exile too, from a hand.
	sorc := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: sorc, Name: "Wrath", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		if err := g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{sorc}, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if !assistCardInZone(g, sorc, game.ZoneExile) {
		t.Error("the discarded sorcery card did not go to exile")
	}
	// A creature card, a land card and a creature that dies still go to a graveyard.
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, bear)
	if !inGraveyardOf(g, me.ID, bear) {
		t.Error("a creature that died was exiled")
	}
}

func TestStdB19DryadMilitantOnlyWorksWhileItIsOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if !assistCardInZone(g, bolt, game.ZoneGraveyard) {
		t.Error("without the Militant the instant is not in the graveyard")
	}
}

// --- Ovika, Enigma Goliath --------------------------------------------

func TestStdB19OvikaMakesGoblinsEqualToTheSpellsManaValueWithTemporaryHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushTestPermanent(g, me.ID, game.Card{Name: "Ovika, Enigma Goliath", OracleID: ovikaEnigmaGoliathOracle,
		TypeLine: "Legendary Creature — Phyrexian Nightmare", Power: 6, Toughness: 6, Colors: []string{"U", "R"}})
	advanceToMain(t, g)
	floatForTest(g, me, "RRR")
	castFromHandForTest(t, g, me, "Shock Wave", "Instant", "{2}{R}", "00000000-0000-0000-0000-0000000000aa",
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	var goblins []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Goblin" && c.Controller == me.ID {
			goblins = append(goblins, c.InstanceID)
		}
	}
	if len(goblins) != 3 {
		t.Fatalf("%d Phyrexian Goblins, want 3 (a mana value 3 spell)", len(goblins))
	}
	for _, id := range goblins {
		if !effectiveAbilitiesContain(t, g, id, "haste") {
			t.Errorf("a Goblin has no haste: %v", effectiveAbilities(t, g, id))
		}
	}
	seq := g.Turn.Seq
	for i := 0; i < 80 && g.Turn.Seq == seq; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if effectiveAbilitiesContain(t, g, goblins[0], "haste") {
		t.Error("the Goblins kept haste past the end of the turn")
	}
}

func TestStdB19OvikaIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushTestPermanent(g, me.ID, game.Card{Name: "Ovika, Enigma Goliath", OracleID: ovikaEnigmaGoliathOracle,
		TypeLine: "Legendary Creature — Phyrexian Nightmare", Power: 6, Toughness: 6, Colors: []string{"U", "R"}})
	advanceToMain(t, g)
	floatForTest(g, me, "RRR")
	castFromHand(t, g, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{R}", Power: 2, Toughness: 2},
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Phyrexian Goblin", me.ID); n != 0 {
		t.Errorf("%d Goblins from a creature spell, want 0", n)
	}
}

// --- Prime Speaker Zegana ---------------------------------------------

func zeganaCast(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	floatForTest(g, me, "GGUUGG")
	id := castFromHand(t, g, game.Card{Name: "Prime Speaker Zegana", TypeLine: "Legendary Creature — Merfolk Wizard",
		ManaCost: "{2}{G}{G}{U}{U}", OracleID: primeSpeakerZeganaOracle, Power: 1, Toughness: 1, Colors: []string{"G", "U"}},
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	return id
}

func TestStdB19ZeganaEntersWithCountersForTheGreatestOtherPowerAndDrawsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	pushTestPermanent(g, me.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 5, Toughness: 4})
	pushTestPermanent(g, opp.ID, game.Card{Name: "Their Dragon", TypeLine: "Creature — Dragon", Power: 9, Toughness: 9})
	hand := me.Hand.Size()
	id := zeganaCast(t, g, me)
	if got := counterCount(g, id, game.CounterPlusOne); got != 5 {
		t.Errorf("%d +1/+1 counters, want 5 (the Ogre's power, not the opposing Dragon's)", got)
	}
	if c := findBattlefieldCardByID(g, id); c == nil || c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("Zegana is %v, want 6/6", c)
	}
	// The cast took Zegana out of hand: net change is the six cards drawn minus her.
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("hand grew by %d, want 6 (the draw equals her power)", got)
	}
}

func TestStdB19ZeganaWithNoOtherCreatureIsAOneOneThatDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id := zeganaCast(t, g, me)
	if got := counterCount(g, id, game.CounterPlusOne); got > 0 {
		t.Errorf("%d counters with no other creature, want none", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
}
