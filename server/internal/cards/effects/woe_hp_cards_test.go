package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_hp_cards_test.go — Wilds of Eldraine, the H–P batch (#4471). Each
// test plays the real catalog entry and asserts what its text says.

const (
	woeHPGluttonOracle     = "aa1b7cdd-0296-418c-ac3b-7ad3c51f1760"
	woeHPSpearguardOracle  = "e3542d0f-fcee-4df7-aa95-b55138caa2ab"
	woeHPNegotiatorOracle  = "90ecd452-333a-43b0-be08-a61f3779b3bd"
	woeHPVigilOracle       = "33fff722-5cc5-4d0e-b788-166b3edb1223"
	woeHPNightmareOracle   = "5e63b69c-d901-4a01-8dfa-88e044856239"
	woeHPChitterlordOracle = "532cac99-e3cf-4869-a19f-cba67a2a6fd0"
	woeHPFlitterfangOracle = "f95bdf0d-5bd1-4215-b143-0930973807ce"
)

func TestWoeHPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeHPGluttonOracle:     "Hamlet Glutton",
		woeHPSpearguardOracle:  "Harried Spearguard",
		woeHPNegotiatorOracle:  "High Fae Negotiator",
		woeHPVigilOracle:       "Hopeful Vigil",
		woeHPNightmareOracle:   "Hopeless Nightmare",
		woeHPChitterlordOracle: "Ogre Chitterlord",
		woeHPFlitterfangOracle: "Old Flitterfang",
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

// --- Hamlet Glutton -------------------------------------------------------------

func TestWoeHPHamletGluttonCostsTwoLessBargained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := woeILPrice(t, g, me.ID, "Hamlet Glutton", "Creature — Giant", "{5}{G}{G}", woeHPGluttonOracle, nil); got != 7 {
		t.Errorf("unbargained price %d, want 7", got)
	}
	if got := woeILPrice(t, g, me.ID, "Hamlet Glutton", "Creature — Giant", "{5}{G}{G}", woeHPGluttonOracle, []int{0}); got != 5 {
		t.Errorf("bargained price %d, want 5", got)
	}
}

func TestWoeHPHamletGluttonBargainedIsCastForFiveAndGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	life := lifeOf(g, me.ID)
	advanceTo(t, g, game.StepPrecombatMain) // mana empties between steps
	floatForTest(g, me, "GGCCC")
	glutton := p1g2Cast(t, g, "Hamlet Glutton", "Creature — Giant", woeHPGluttonOracle, "{5}{G}{G}", game.CastSpellParams{
		Strict: true, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(glutton) {
		t.Fatal("Hamlet Glutton did not resolve")
	}
	assertKeywords(t, g, glutton, "trample")
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("you are at %d, want %d", got, life+3)
	}
}

// --- Harried Spearguard ----------------------------------------------------------

func TestWoeHPHarriedSpearguardHasHasteAndLeavesARatWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	guard := b12Push(g, me.ID, "Harried Spearguard", "Creature — Human Soldier", woeHPSpearguardOracle, 1, 1)
	assertKeywords(t, g, guard, "haste")
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("setup: %d Rats", got)
	}
	woeB1Destroy(t, g, guard)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rat tokens after it died, want 1", got)
	}
	assertRestrictions(t, g, findBattlefieldByName(g, "Rat"), game.CantBlock)
}

// --- High Fae Negotiator ---------------------------------------------------------

func TestWoeHPHighFaeNegotiatorDrainsThreeOnlyWhenBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	life := lifeOf(g, me.ID)
	opps := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps[p.ID] = lifeOf(g, p.ID)
		}
	}
	fae := p1g2Cast(t, g, "High Fae Negotiator", "Creature — Faerie Warlock", woeHPNegotiatorOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fae) {
		t.Fatal("High Fae Negotiator did not resolve")
	}
	assertKeywords(t, g, fae, "flying")
	for id, before := range opps {
		if got := lifeOf(g, id); got != before-3 {
			t.Errorf("an opponent is at %d, want %d", got, before-3)
		}
	}
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("you are at %d, want %d", got, life+3)
	}
}

func TestWoeHPHighFaeNegotiatorUnbargainedDoesNothingOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	life, theirs := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	fae := castCatalogSpell(t, g, "High Fae Negotiator", "Creature — Faerie Warlock", woeHPNegotiatorOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fae) {
		t.Fatal("High Fae Negotiator did not resolve")
	}
	if lifeOf(g, me.ID) != life || lifeOf(g, opp.ID) != theirs {
		t.Errorf("unbargained, life moved: you %d → %d, opponent %d → %d", life, lifeOf(g, me.ID), theirs, lifeOf(g, opp.ID))
	}
}

// --- Hopeful Vigil ---------------------------------------------------------------

func TestWoeHPHopefulVigilMakesAKnightAndScriesTwoWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	vigil := castCatalogSpell(t, g, "Hopeful Vigil", "Enchantment", woeHPVigilOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Knight"); got != 1 {
		t.Fatalf("%d Knight tokens after it entered, want 1", got)
	}
	knight := findBattlefieldByName(g, "Knight")
	if p, tough := effectivePTOf(t, g, knight); p != 2 || tough != 2 {
		t.Errorf("the Knight is %d/%d, want 2/2", p, tough)
	}
	assertKeywords(t, g, knight, "vigilance")

	if err := g.ActivateCatalogAbility(me.ID, vigil, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the sacrifice ability activated without {2}{W}")
	}
	floatForTest(g, me, "WCC")
	if err := g.ActivateCatalogAbility(me.ID, vigil, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{W}: %v", err)
	}
	if !g.Battlefield.Contains(vigil) {
		t.Fatal("the enchantment was sacrificed as a cost, not as the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vigil) {
		t.Fatal("the ability did not sacrifice the enchantment")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Hopeless Nightmare ----------------------------------------------------------

func TestWoeHPHopelessNightmareMakesEachOpponentDiscardAndLoseTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			before[p.ID] = lifeOf(g, p.ID)
		}
	}
	castCatalogSpell(t, g, "Hopeless Nightmare", "Enchantment", woeHPNightmareOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if owed := discardOwed(g, me.ID); owed != 0 {
		t.Errorf("you owe %d discards, want 0", owed)
	}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if owed := discardOwed(g, p.ID); owed != 1 {
			t.Errorf("an opponent owes %d discards, want 1", owed)
		}
		discardFromHand(t, g, p.ID)
	}
	passPriorityAroundTable(t, g)
	for id, life := range before {
		if got := lifeOf(g, id); got != life-2 {
			t.Errorf("an opponent is at %d, want %d", got, life-2)
		}
	}
}

func TestWoeHPHopelessNightmareScriesTwoWhenDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	nightmare := pushCatalogPermanent(g, me.ID, "Hopeless Nightmare", "Enchantment", woeHPNightmareOracle, false)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(nightmare); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Ogre Chitterlord ------------------------------------------------------------

func TestWoeHPOgreChitterlordMakesTwoRatsAndPumpsOnlyAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	ogre := castCatalogSpell(t, g, "Ogre Chitterlord", "Creature — Ogre Warrior", woeHPChitterlordOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Fatal("Ogre Chitterlord did not resolve")
	}
	assertKeywords(t, g, ogre, "menace")
	if got := b27CountNamed(g, me.ID, "Rat"); got != 2 {
		t.Fatalf("%d Rats after it entered, want 2", got)
	}
	rat := findBattlefieldByName(g, "Rat")
	assertRestrictions(t, g, rat, game.CantBlock)
	if p := effectivePower(t, g, rat); p != 1 {
		t.Errorf("with two Rats a Rat has power %d, want 1", p)
	}
}

func TestWoeHPOgreChitterlordAttackPumpsEachRatAtFiveOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, CantBlockRatToken(), 3) })
	theirRat := b12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	ogre := b12Push(g, me.ID, "Ogre Chitterlord", "Creature — Ogre Warrior", woeHPChitterlordOracle, 6, 5)
	woeMPReady(g, ogre)
	declareAttack(t, g, opp.ID, ogre)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 5 {
		t.Fatalf("%d Rats after the attack trigger, want 5", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Rat" {
			if p := effectivePower(t, g, c.InstanceID); p != 3 {
				t.Errorf("a Rat has power %d with five Rats, want 3", p)
			}
		}
	}
	if p := effectivePower(t, g, ogre); p != 6 {
		t.Errorf("the Ogre, not a Rat, has power %d, want 6", p)
	}
	if p := effectivePower(t, g, theirRat); p != 1 {
		t.Errorf("an opponent's Rat has power %d, want 1", p)
	}
}

// --- Old Flitterfang -------------------------------------------------------------

func TestWoeHPOldFlitterfangMakesFoodAtAnEndStepOnlyAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	fang := b12Push(g, me.ID, "Old Flitterfang", "Legendary Creature — Rat Faerie", woeHPFlitterfangOracle, 3, 4)
	assertKeywords(t, g, fang, "flying")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food at an end step with no death, want 0", got)
	}
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	other := g.Seats[g.Turn.ActiveSeat]
	victim := b12Creature(g, other.ID, "Their Bear", woeMPBearType, 2, 2)
	woeB1Destroy(t, g, victim)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food at another player's end step after a death, want 1", got)
	}
}

func TestWoeHPOldFlitterfangSacrificesAnotherCreatureOrArtifactToGrow(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	fang := b12Push(g, me.ID, "Old Flitterfang", "Legendary Creature — Rat Faerie", woeHPFlitterfangOracle, 3, 4)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	advanceTo(t, g, game.StepPrecombatMain)
	for name, id := range map[string]uuid.UUID{"itself": fang, "a land": land} {
		floatForTest(g, me, "BCC")
		if err := g.ActivateCatalogAbility(me.ID, fang, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("Old Flitterfang sacrificed %s", name)
		}
	}
	floatForTest(g, me, "BCC")
	if err := g.ActivateCatalogAbility(me.ID, fang, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("{2}{B}, sacrifice an artifact: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, fang); p != 5 || tough != 6 {
		t.Errorf("Old Flitterfang is %d/%d, want 5/6", p, tough)
	}
}
