package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b03_cards_test.go — S82 AFR batch 3 (#5499): one-shot instants
// and sorceries. Every test plays the real catalog card through the engine.

const (
	afrB03BullsStrength = "a5f10b54-10db-4c92-a596-8afea4e470f5"
	afrB03Eyes          = "0fae0474-25ec-474f-8c7f-9e9061835030"
	afrB03ShockingGrasp = "2ebe4c68-87eb-474d-a4c4-0468835b4900"
	afrB03Improvised    = "fa5dad0b-8e23-4b88-b383-849f7318e01f"
	afrB03GrimBounty    = "8060dc97-9868-4846-85f4-0ef70da2b021"
	afrB03Clutches      = "0bbab99f-2ff6-4d48-b7c1-1ab0500c9422"
	afrB03PowerWordKill = "ea935885-acbc-450d-a3ad-f80c09e12ce7"
	afrB03Rally         = "6bbdf771-75d5-4910-bd16-6232e254f7f6"
	afrB03BurningHands  = "51a0d1c2-ff29-4ac9-84a2-ad4f566fee8d"
	afrB03Ray           = "647e5639-6339-48af-af96-e48729e1cad3"
	afrB03Polymorph     = "5ac8c234-030d-4ba1-bc2f-d05761045278"
	afrB03SuddenInsight = "8e0b605f-2095-4189-8c71-9d1ac93db42e"
	afrB03CheckForTraps = "41b2ea88-b3b9-44f0-8886-7b2f37b22ebf"
	afrB03TreasureName  = "Treasure"
)

func TestAfrB03EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB03BullsStrength: "Bull's Strength", afrB03Eyes: "Eyes of the Beholder", afrB03ShockingGrasp: "Shocking Grasp",
		afrB03Improvised: "Improvised Weaponry", afrB03GrimBounty: "Grim Bounty", afrB03Clutches: "Demogorgon's Clutches",
		afrB03PowerWordKill: "Power Word Kill", afrB03Rally: "Rally Maneuver", afrB03BurningHands: "Burning Hands",
		afrB03Ray: "Ray of Enfeeblement", afrB03Polymorph: "Mordenkainen's Polymorph", afrB03SuddenInsight: "Sudden Insight",
		afrB03CheckForTraps: "Check for Traps",
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

func afrB03Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, oracle, targets)
	passPriorityAroundTable(t, g)
}

func afrB03Colored(g *game.Game, owner uuid.UUID, name string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func afrB03Treasures(g *game.Game, p uuid.UUID) (n int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == afrB03TreasureName && c.Controller == p {
				n++
			}
		}
	})
	return n
}

func afrB03OnBattlefield(g *game.Game, id uuid.UUID) (on bool) {
	g.ReadSnapshot(func() { on = g.Battlefield.Contains(id) })
	return on
}

func TestAfrB03BullsStrengthPumpsGrantsTrampleAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := afrB03Colored(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	afrB03Cast(t, g, "Bull's Strength", "Instant", afrB03BullsStrength, pbCard(bear))
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveAbility(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear && c.Tapped {
				t.Error("the Bear is still tapped")
			}
		}
	})
	advanceToNextSeatsTurn(t, g)
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("P/T after cleanup = %d/%d, want 2/2", p, tt)
	}
	if hasEffectiveAbility(t, g, bear, "trample") {
		t.Error("trample outlived the turn")
	}
}

func TestAfrB03BullsStrengthNeedsACreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	land := b12Permanent(g, g.Seats[0].ID, "Rock", "Artifact")
	err := castCatalogSpellErrStrict(t, g, "Bull's Strength", "Instant", afrB03BullsStrength, []game.TargetRef{pbCard(land)})
	if err == nil {
		t.Error("Bull's Strength accepted a noncreature target")
	}
}

func TestAfrB03EyesOfTheBeholderKillsAndSizes(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	giant := afrB03Colored(g, opp.ID, "Giant", 12, 12)
	afrB03Cast(t, g, "Eyes of the Beholder", "Instant", afrB03Eyes, pbCard(giant))
	if p, tt := ptOf(t, g, giant); p != 1 || tt != 1 {
		t.Errorf("Giant is %d/%d, want 1/1", p, tt)
	}
	bear := afrB03Colored(g, opp.ID, "Bear", 2, 2)
	afrB03Cast(t, g, "Eyes of the Beholder", "Instant", afrB03Eyes, pbCard(bear))
	if afrB03OnBattlefield(g, bear) {
		t.Error("the 2/2 survived -11/-11")
	}
}

func TestAfrB03ShockingGraspShrinksAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := afrB03Colored(g, opp.ID, "Bear", 3, 3)
	before := me.Hand.Size()
	afrB03Cast(t, g, "Shocking Grasp", "Instant", afrB03ShockingGrasp, pbCard(bear))
	if p, tt := ptOf(t, g, bear); p != 1 || tt != 3 {
		t.Errorf("Bear is %d/%d, want 1/3", p, tt)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (the spell left the hand, one card drawn)", got, before+1)
	}
}

func TestAfrB03ImprovisedWeaponryDamagesAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	afrB03Cast(t, g, "Improvised Weaponry", "Sorcery", afrB03Improvised, pbPlayer(opp.ID))
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
	if n := afrB03Treasures(g, me.ID); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestAfrB03GrimBountyDestroysAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := afrB03Colored(g, opp.ID, "Bear", 2, 2)
	afrB03Cast(t, g, "Grim Bounty", "Sorcery", afrB03GrimBounty, pbCard(bear))
	if afrB03OnBattlefield(g, bear) {
		t.Error("the Bear survived")
	}
	if n := afrB03Treasures(g, me.ID); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

func TestAfrB03GrimBountyRefusesAPlayerTarget(t *testing.T) {
	g := newCatalogGame(t)
	err := castCatalogSpellErrStrict(t, g, "Grim Bounty", "Sorcery", afrB03GrimBounty, []game.TargetRef{pbPlayer(g.Seats[1].ID)})
	if err == nil {
		t.Error("Grim Bounty accepted a player")
	}
}

func TestAfrB03PowerWordKillSparesAngelsDemonsDevilsAndDragons(t *testing.T) {
	for _, sub := range []string{"Angel", "Demon", "Devil", "Dragon"} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		id := b12Creature(g, opp.ID, sub, "Creature — "+sub, 3, 3)
		err := castCatalogSpellErrStrict(t, g, "Power Word Kill", "Instant", afrB03PowerWordKill, []game.TargetRef{pbCard(id)})
		if err == nil {
			t.Errorf("Power Word Kill accepted a %s", sub)
		}
	}
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	afrB03Cast(t, g, "Power Word Kill", "Instant", afrB03PowerWordKill, pbCard(bear))
	if afrB03OnBattlefield(g, bear) {
		t.Error("the Bear survived")
	}
}

func TestAfrB03RallyManeuverGivesEachCreatureItsHalf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := afrB03Colored(g, me.ID, "A", 2, 2)
	b := afrB03Colored(g, me.ID, "B", 1, 1)
	afrB03Cast(t, g, "Rally Maneuver", "Instant", afrB03Rally,
		game.TargetRef{Kind: game.TargetCard, ID: a, Slot: 0}, game.TargetRef{Kind: game.TargetCard, ID: b, Slot: 1})
	if p, tt := ptOf(t, g, a); p != 4 || tt != 2 {
		t.Errorf("A is %d/%d, want 4/2", p, tt)
	}
	if p, tt := ptOf(t, g, b); p != 1 || tt != 3 {
		t.Errorf("B is %d/%d, want 1/3", p, tt)
	}
	if !hasEffectiveAbility(t, g, a, "first strike") || hasEffectiveAbility(t, g, a, "lifelink") {
		t.Error("A should have first strike only")
	}
	if !hasEffectiveAbility(t, g, b, "lifelink") || hasEffectiveAbility(t, g, b, "first strike") {
		t.Error("B should have lifelink only")
	}
}

func TestAfrB03RallyManeuverSecondTargetIsOptionalAndMustDiffer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := afrB03Colored(g, me.ID, "A", 2, 2)
	afrB03Cast(t, g, "Rally Maneuver", "Instant", afrB03Rally, game.TargetRef{Kind: game.TargetCard, ID: a, Slot: 0})
	if p, tt := ptOf(t, g, a); p != 4 || tt != 2 {
		t.Errorf("A is %d/%d, want 4/2", p, tt)
	}
	g2 := newCatalogGame(t)
	same := afrB03Colored(g2, g2.Seats[0].ID, "A", 2, 2)
	err := castCatalogSpellErrStrict(t, g2, "Rally Maneuver", "Instant", afrB03Rally,
		[]game.TargetRef{{Kind: game.TargetCard, ID: same, Slot: 0}, {Kind: game.TargetCard, ID: same, Slot: 1}})
	if err == nil {
		t.Error("Rally Maneuver accepted the same creature for both targets")
	}
}

func TestAfrB03BurningHandsDamageDependsOnGreen(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	green := afrB03Colored(g, opp.ID, "Green Beast", 5, 6, "G")
	afrB03Cast(t, g, "Burning Hands", "Instant", afrB03BurningHands, pbCard(green))
	if afrB03OnBattlefield(g, green) {
		t.Error("the 5/6 green creature survived 6 damage")
	}
	red := afrB03Colored(g, opp.ID, "Red Beast", 5, 6, "R")
	afrB03Cast(t, g, "Burning Hands", "Instant", afrB03BurningHands, pbCard(red))
	if !afrB03OnBattlefield(g, red) {
		t.Error("the non-green creature died to 2 damage")
	}
}

func TestAfrB03BurningHandsHitsAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pw := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker — Test", Counters: map[string]int{"loyalty": 5},
		Owner: opp.ID, Controller: opp.ID,
	})
	afrB03Cast(t, g, "Burning Hands", "Instant", afrB03BurningHands, pbCard(pw))
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == pw && c.Counters["loyalty"] != 3 {
				t.Errorf("loyalty %d, want 3", c.Counters["loyalty"])
			}
		}
	})
}

func TestAfrB03RayOfEnfeeblementWhiteGetsMinusFourFour(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	white := afrB03Colored(g, opp.ID, "Angel Knight", 4, 4, "W")
	afrB03Cast(t, g, "Ray of Enfeeblement", "Instant", afrB03Ray, pbCard(white))
	if afrB03OnBattlefield(g, white) {
		t.Error("the white 4/4 survived -4/-4")
	}
	blue := afrB03Colored(g, opp.ID, "Blue Beast", 4, 4, "U")
	afrB03Cast(t, g, "Ray of Enfeeblement", "Instant", afrB03Ray, pbCard(blue))
	if p, tt := ptOf(t, g, blue); p != 0 || tt != 3 {
		t.Errorf("blue creature is %d/%d, want 0/3", p, tt)
	}
}

func TestAfrB03PolymorphMakesA44FlyingDragonUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear Warrior", 2, 2)
	afrB03Cast(t, g, "Mordenkainen's Polymorph", "Instant", afrB03Polymorph, pbCard(bear))
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("P/T %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveAbility(t, g, bear, "flying") {
		t.Error("no flying")
	}
	var typeLine []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bear {
				typeLine = c.Effective().Subtypes
			}
		}
	})
	if len(typeLine) != 1 || typeLine[0] != "Dragon" {
		t.Errorf("creature types %v, want only Dragon", typeLine)
	}
	advanceToNextSeatsTurn(t, g)
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("after cleanup %d/%d, want 2/2", p, tt)
	}
	if hasEffectiveAbility(t, g, bear, "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestAfrB03SuddenInsightDrawsPerDifferentManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b17GraveyardCard(me, "One A", "Creature — Test", "{1}")
	b17GraveyardCard(me, "One B", "Instant", "{W}")
	b17GraveyardCard(me, "Three", "Sorcery", "{2}{R}")
	b17GraveyardCard(me, "Zero", "Artifact", "")
	b17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	b17GraveyardCard(me, "Other Land", "Land", "{5}")
	before := me.Hand.Size()
	afrB03Cast(t, g, "Sudden Insight", "Instant", afrB03SuddenInsight)
	// The spell enters and leaves the hand; 3 different values (0, 1, 3) are drawn.
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d", got, before+3)
	}
}

func TestAfrB03SuddenInsightWithAnEmptyGraveyardDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	afrB03Cast(t, g, "Sudden Insight", "Instant", afrB03SuddenInsight)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d", got, before)
	}
}

func TestAfrB03DemogorgonsClutchesDiscardsMillsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = me
	hand := revealHand(opp, rhForest(), rhBolt(), rhFireball())
	life := opp.Life
	lib := opp.Library.Size()
	castCatalogSpell(t, g, "Demogorgon's Clutches", "Sorcery", afrB03Clutches, []game.TargetRef{pbPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, hand[0], hand[1])
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hand[0]) || !opp.Graveyard.Contains(hand[1]) || opp.Hand.Size() != 1 {
		t.Errorf("the opponent did not discard the two cards they chose (hand %d)", opp.Hand.Size())
	}
	if got := opp.Library.Size(); got != lib-2 {
		t.Errorf("library %d, want %d", got, lib-2)
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if opp.Graveyard.Size() != 4 {
		t.Errorf("graveyard %d, want 4 (2 discarded + 2 milled)", opp.Graveyard.Size())
	}
}

func TestAfrB03DemogorgonsClutchesRefusesYourself(t *testing.T) {
	g := newCatalogGame(t)
	err := castCatalogSpellErrStrict(t, g, "Demogorgon's Clutches", "Sorcery", afrB03Clutches, []game.TargetRef{pbPlayer(g.Seats[0].ID)})
	if err == nil {
		t.Error("Demogorgon's Clutches targeted its caster")
	}
}

func afrB03RunCheckForTraps(t *testing.T, cards ...game.Card) (g *game.Game, lifeMe, lifeOpp int, ids []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids = revealHand(opp, cards...)
	castCatalogSpell(t, g, "Check for Traps", "Sorcery", afrB03CheckForTraps, []game.TargetRef{pbPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	lifeMe, lifeOpp = me.Life, opp.Life
	if pick := latestOpenVariant(g); pick != nil {
		if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[len(ids)-1]}); err != nil {
			t.Fatalf("resolve: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	return g, lifeMe, lifeOpp, ids
}

func latestOpenVariant(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealedHandPick {
			return c
		}
	}
	return nil
}

func TestAfrB03CheckForTrapsInstantCostsTheOpponentLife(t *testing.T) {
	g, _, _, ids := afrB03RunCheckForTraps(t, rhForest(), rhBolt())
	me, opp := g.Seats[0], g.Seats[1]
	if !g.Exile.Contains(ids[1]) {
		t.Fatal("the instant was not exiled")
	}
	if opp.Life != 39 || me.Life != 40 {
		t.Errorf("life me %d opp %d, want 40 and 39", me.Life, opp.Life)
	}
}

func TestAfrB03CheckForTrapsFlashCreatureCostsTheOpponentLife(t *testing.T) {
	flash := game.Card{Name: "Flash Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Keywords: []string{"flash"}}
	g, _, _, ids := afrB03RunCheckForTraps(t, rhForest(), flash)
	me, opp := g.Seats[0], g.Seats[1]
	if !g.Exile.Contains(ids[1]) {
		t.Fatal("the flash creature was not exiled")
	}
	if opp.Life != 39 || me.Life != 40 {
		t.Errorf("life me %d opp %d, want 40 and 39", me.Life, opp.Life)
	}
}

func TestAfrB03CheckForTrapsOtherwiseCostsYou(t *testing.T) {
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	g, _, _, ids := afrB03RunCheckForTraps(t, rhForest(), bear)
	me, opp := g.Seats[0], g.Seats[1]
	if !g.Exile.Contains(ids[1]) {
		t.Fatal("the creature was not exiled")
	}
	if me.Life != 39 || opp.Life != 40 {
		t.Errorf("life me %d opp %d, want 39 and 40", me.Life, opp.Life)
	}
}

func TestAfrB03CheckForTrapsWithOnlyLandsCostsYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	revealHand(opp, rhForest())
	castCatalogSpell(t, g, "Check for Traps", "Sorcery", afrB03CheckForTraps, []game.TargetRef{pbPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if me.Life != 39 || opp.Life != 40 {
		t.Errorf("life me %d opp %d, want 39 and 40 (nothing was exiled)", me.Life, opp.Life)
	}
}
