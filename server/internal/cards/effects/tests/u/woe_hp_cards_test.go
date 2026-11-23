package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Hamlet Glutton -------------------------------------------------------------

func TestWoeHPHamletGluttonCostsTwoLessBargained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := woeILPrice(t, g, me.ID, "Hamlet Glutton", "Creature — Giant", "{5}{G}{G}", woeHPGluttonOracle, nil); got != 7 {
		t.Errorf("unbargained price %d, want 7", got)
	}
	if got := woeILPrice(t, g, me.ID, "Hamlet Glutton", "Creature — Giant", "{5}{G}{G}", woeHPGluttonOracle, []int{0}); got != 5 {
		t.Errorf("bargained price %d, want 5", got)
	}
}

func TestWoeHPHamletGluttonBargainedIsCastForFiveAndGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	token := woeILToken(g, me.ID)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain) // mana empties between steps
	effectstest.FloatForTest(g, me, "GGCCC")
	glutton := effectstest.P1g2Cast(t, g, "Hamlet Glutton", "Creature — Giant", woeHPGluttonOracle, "{5}{G}{G}", game.CastSpellParams{
		Strict: true, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(glutton) {
		t.Fatal("Hamlet Glutton did not resolve")
	}
	effectstest.AssertKeywords(t, g, glutton, "trample")
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("you are at %d, want %d", got, life+3)
	}
}

// --- Harried Spearguard ----------------------------------------------------------

func TestWoeHPHarriedSpearguardHasHasteAndLeavesARatWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	guard := effectstest.B12Push(g, me.ID, "Harried Spearguard", "Creature — Human Soldier", woeHPSpearguardOracle, 1, 1)
	effectstest.AssertKeywords(t, g, guard, "haste")
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("setup: %d Rats", got)
	}
	effectstest.WoeB1Destroy(t, g, guard)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rat tokens after it died, want 1", got)
	}
	effectstest.AssertRestrictions(t, g, effectstest.FindBattlefieldByName(g, "Rat"), game.CantBlock)
}

// --- High Fae Negotiator ---------------------------------------------------------

func TestWoeHPHighFaeNegotiatorDrainsThreeOnlyWhenBargained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	token := woeILToken(g, me.ID)
	life := effectstest.LifeOf(g, me.ID)
	opps := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps[p.ID] = effectstest.LifeOf(g, p.ID)
		}
	}
	fae := effectstest.P1g2Cast(t, g, "High Fae Negotiator", "Creature — Faerie Warlock", woeHPNegotiatorOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fae) {
		t.Fatal("High Fae Negotiator did not resolve")
	}
	effectstest.AssertKeywords(t, g, fae, "flying")
	for id, before := range opps {
		if got := effectstest.LifeOf(g, id); got != before-3 {
			t.Errorf("an opponent is at %d, want %d", got, before-3)
		}
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("you are at %d, want %d", got, life+3)
	}
}

func TestWoeHPHighFaeNegotiatorUnbargainedDoesNothingOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	life, theirs := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	fae := effectstest.CastCatalogSpell(t, g, "High Fae Negotiator", "Creature — Faerie Warlock", woeHPNegotiatorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fae) {
		t.Fatal("High Fae Negotiator did not resolve")
	}
	if effectstest.LifeOf(g, me.ID) != life || effectstest.LifeOf(g, opp.ID) != theirs {
		t.Errorf("unbargained, life moved: you %d → %d, opponent %d → %d", life, effectstest.LifeOf(g, me.ID), theirs, effectstest.LifeOf(g, opp.ID))
	}
}

// --- Hopeful Vigil ---------------------------------------------------------------

func TestWoeHPHopefulVigilMakesAKnightAndScriesTwoWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	vigil := effectstest.CastCatalogSpell(t, g, "Hopeful Vigil", "Enchantment", woeHPVigilOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Knight"); got != 1 {
		t.Fatalf("%d Knight tokens after it entered, want 1", got)
	}
	knight := effectstest.FindBattlefieldByName(g, "Knight")
	if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 2 || tough != 2 {
		t.Errorf("the Knight is %d/%d, want 2/2", p, tough)
	}
	effectstest.AssertKeywords(t, g, knight, "vigilance")

	if err := g.ActivateCatalogAbility(me.ID, vigil, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the sacrifice ability activated without {2}{W}")
	}
	effectstest.FloatForTest(g, me, "WCC")
	if err := g.ActivateCatalogAbility(me.ID, vigil, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{W}: %v", err)
	}
	if !g.Battlefield.Contains(vigil) {
		t.Fatal("the enchantment was sacrificed as a cost, not as the ability resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vigil) {
		t.Fatal("the ability did not sacrifice the enchantment")
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

// --- Hopeless Nightmare ----------------------------------------------------------

func TestWoeHPHopelessNightmareMakesEachOpponentDiscardAndLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			before[p.ID] = effectstest.LifeOf(g, p.ID)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Hopeless Nightmare", "Enchantment", woeHPNightmareOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 0 {
		t.Errorf("you owe %d discards, want 0", owed)
	}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if owed := effectstest.DiscardOwed(g, p.ID); owed != 1 {
			t.Errorf("an opponent owes %d discards, want 1", owed)
		}
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, life := range before {
		if got := effectstest.LifeOf(g, id); got != life-2 {
			t.Errorf("an opponent is at %d, want %d", got, life-2)
		}
	}
}

func TestWoeHPHopelessNightmareScriesTwoWhenDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	nightmare := effectstest.PushCatalogPermanent(g, me.ID, "Hopeless Nightmare", "Enchantment", woeHPNightmareOracle, false)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(nightmare); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

// --- Ogre Chitterlord ------------------------------------------------------------

func TestWoeHPOgreChitterlordMakesTwoRatsAndPumpsOnlyAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	ogre := effectstest.CastCatalogSpell(t, g, "Ogre Chitterlord", "Creature — Ogre Warrior", woeHPChitterlordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Fatal("Ogre Chitterlord did not resolve")
	}
	effectstest.AssertKeywords(t, g, ogre, "menace")
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 2 {
		t.Fatalf("%d Rats after it entered, want 2", got)
	}
	rat := effectstest.FindBattlefieldByName(g, "Rat")
	effectstest.AssertRestrictions(t, g, rat, game.CantBlock)
	if p := effectstest.EffectivePower(t, g, rat); p != 1 {
		t.Errorf("with two Rats a Rat has power %d, want 1", p)
	}
}

func TestWoeHPOgreChitterlordAttackPumpsEachRatAtFiveOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.CantBlockRatToken(), 3) })
	theirRat := effectstest.B12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	ogre := effectstest.B12Push(g, me.ID, "Ogre Chitterlord", "Creature — Ogre Warrior", woeHPChitterlordOracle, 6, 5)
	woeMPReady(g, ogre)
	effectstest.DeclareAttack(t, g, opp.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 5 {
		t.Fatalf("%d Rats after the attack trigger, want 5", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Rat" {
			if p := effectstest.EffectivePower(t, g, c.InstanceID); p != 3 {
				t.Errorf("a Rat has power %d with five Rats, want 3", p)
			}
		}
	}
	if p := effectstest.EffectivePower(t, g, ogre); p != 6 {
		t.Errorf("the Ogre, not a Rat, has power %d, want 6", p)
	}
	if p := effectstest.EffectivePower(t, g, theirRat); p != 1 {
		t.Errorf("an opponent's Rat has power %d, want 1", p)
	}
}

// --- Old Flitterfang -------------------------------------------------------------

func TestWoeHPOldFlitterfangMakesFoodAtAnEndStepOnlyAfterADeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	fang := effectstest.B12Push(g, me.ID, "Old Flitterfang", "Legendary Creature — Rat Faerie", woeHPFlitterfangOracle, 3, 4)
	effectstest.AssertKeywords(t, g, fang, "flying")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food at an end step with no death, want 0", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	other := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.B12Creature(g, other.ID, "Their Bear", woeMPBearType, 2, 2)
	effectstest.WoeB1Destroy(t, g, victim)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food at another player's end step after a death, want 1", got)
	}
}

func TestWoeHPOldFlitterfangSacrificesAnotherCreatureOrArtifactToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	fang := effectstest.B12Push(g, me.ID, "Old Flitterfang", "Legendary Creature — Rat Faerie", woeHPFlitterfangOracle, 3, 4)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	for name, id := range map[string]uuid.UUID{"itself": fang, "a land": land} {
		effectstest.FloatForTest(g, me, "BCC")
		if err := g.ActivateCatalogAbility(me.ID, fang, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("Old Flitterfang sacrificed %s", name)
		}
	}
	effectstest.FloatForTest(g, me, "BCC")
	if err := g.ActivateCatalogAbility(me.ID, fang, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("{2}{B}, sacrifice an artifact: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, fang); p != 5 || tough != 6 {
		t.Errorf("Old Flitterfang is %d/%d, want 5/6", p, tough)
	}
}
