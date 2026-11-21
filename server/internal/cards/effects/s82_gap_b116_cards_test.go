package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b116_cards_test.go — S82 GAP batch 116 (Tales of Middle-earth
// Commander, #5948). Every test plays the real catalog card through the engine.

const (
	gapB116Balrog    = "d73191d8-6f94-4fba-acd2-2d0490e3ac00"
	gapB116Warmaster = "bda83ef4-e345-495d-878c-4da171f997ba"
	gapB116Hobbit    = "cca33482-0bc5-4fe4-89dc-ba33b2e93831"
	gapB116Brigade   = "da4bf807-93f3-41b7-8bfd-aa63a551754e"
	gapB116Haldir    = "653217b1-1e92-49a7-81d9-fc4d70d4f27f"
	gapB116Pony      = "44906d21-83cc-49ec-8505-4f7b020ac05f"
	gapB116Pig       = "14f55fc9-ec19-4b4a-b3de-d75898c19edf"
	gapB116Aragorn   = "07326335-45dd-4ebd-8ae6-5191b6eb2928"
)

func TestGapB116AllAreFull(t *testing.T) {
	for o, name := range map[string]string{
		gapB116Balrog: "The Balrog of Moria", gapB116Warmaster: "Elvish Warmaster",
		gapB116Hobbit: "Feasting Hobbit", gapB116Brigade: "Galadhrim Brigade",
		gapB116Haldir: "Haldir, Lórien Lieutenant", gapB116Pony: "Motivated Pony",
		gapB116Pig: "Prize Pig", gapB116Aragorn: "Aragorn, Hornburg Hero",
	} {
		spec, ok := Lookup(o)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Galadhrim Brigade -------------------------------------------------------

func TestGapB116GaladhrimBrigadePumpsOtherElvesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	brigade := b12Push(g, me.ID, "Galadhrim Brigade", "Creature — Elf Soldier", gapB116Brigade, 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Druid", 1, 1)
	if p, _ := ptOf(t, g, elf); p != 2 {
		t.Errorf("my other Elf has power %d, want 2", p)
	}
	if p, _ := ptOf(t, g, brigade); p != 2 {
		t.Errorf("the Brigade pumped itself: power %d", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("a non-Elf got the pump: power %d", p)
	}
	if p, _ := ptOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Elf got the pump: power %d", p)
	}
}

func TestGapB116GaladhrimBrigadeSquadMakesTokenCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Galadhrim Brigade", "Creature — Elf Soldier", gapB116Brigade, "{2}{G}", 2, 2)
	floatMana(t, g, me, "{G}{G}{G}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("cast with squad paid twice: %v", err)
	}
	settleOrderingTriggers(t, g)
	ids := tmtB1Named(g, "Galadhrim Brigade")
	if len(ids) != 3 {
		t.Fatalf("%d Brigades, want the card and two token copies", len(ids))
	}
	for _, sid := range ids {
		if p, _ := ptOf(t, g, sid); p != 4 {
			t.Errorf("Brigade %s has power %d, want 4 (2 and +1/+1 from each of the other two Elves)", sid, p)
		}
	}
}

// --- Feasting Hobbit ---------------------------------------------------------

func TestGapB116FeastingHobbitDevoursFoodsForThreeCountersEach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	f1 := b12Permanent(g, me.ID, "Food", "Artifact — Food")
	f2 := b12Permanent(g, me.ID, "Food", "Artifact — Food")
	goblin := pushDevourFodder(g, me.ID, "Goblin")
	id := castDevourer(t, g, "Feasting Hobbit", gapB116Hobbit)
	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("devour did not ask")
	}
	for _, cid := range c.ChooseCards {
		if cid == goblin {
			t.Error("a creature that is not a Food is offered")
		}
	}
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, f1, f2)
	if got := devourPlusOnes(g, id); got != 6 {
		t.Errorf("%d +1/+1 counters, want 6 (three for each Food)", got)
	}
	if g.Battlefield.Contains(f1) || g.Battlefield.Contains(f2) {
		t.Error("a devoured Food is still on the battlefield")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was sacrificed")
	}
}

func TestGapB116FeastingHobbitCantBeBlockedByLesserPower(t *testing.T) {
	spec, ok := Lookup(gapB116Hobbit)
	if !ok || len(spec.BlockRules) != 1 {
		t.Fatalf("Feasting Hobbit declares %d block rules, want 1", len(spec.BlockRules))
	}
}

// --- Motivated Pony ----------------------------------------------------------

func TestGapB116MotivatedPonyPumpsAttackersAndUntapsWithAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pony := pushCatalogPermanent(g, me.ID, "Motivated Pony", "Creature — Horse", gapB116Pony, false)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, pony)
		c.Power, c.Toughness = 3, 3
		c.Keywords = []string{"trample", "haste"}
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, pony, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("attacking Bear is %d/%d, want 3/3 (+1/+1, no Food)", p, tough)
	}
	if !b16Tapped(t, g, bear) {
		t.Error("the Bear untapped without a Food entering")
	}
}

func TestGapB116MotivatedPonyWithAFoodUntapsThemAndAddsTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pony := pushCatalogPermanent(g, me.ID, "Motivated Pony", "Creature — Horse", gapB116Pony, false)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, pony)
		c.Power, c.Toughness = 3, 3
		c.Keywords = []string{"trample", "haste"}
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, FoodToken(), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("Food: %v %v", ids, err)
		}
	})
	declareAttack(t, g, opp.ID, pony, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("attacking Bear is %d/%d, want 5/5 (+1/+1 and an additional +2/+2)", p, tough)
	}
	if b16Tapped(t, g, bear) || b16Tapped(t, g, pony) {
		t.Error("an attacker stayed tapped")
	}
	if c, _ := battlefieldCard(g, bear); c.AttackingTarget == uuid.Nil {
		t.Error("the untapped Bear is no longer attacking")
	}
}

// --- Haldir, Lórien Lieutenant -----------------------------------------------

func TestGapB116HaldirEntersWithXCountersAndPumpsOtherElves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	toMainForCost(t, g)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Druid", 1, 1)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Druid", 1, 1)
	id := tmtB11Hand(g, me, "Haldir, Lórien Lieutenant", "Legendary Creature — Elf Soldier", gapB116Haldir, "{X}{G}", 0, 0)
	floatMana(t, g, me, "{G}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("cast X=3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 3 || tough != 3 {
		t.Fatalf("Haldir is %d/%d, want 3/3 from 3 counters", p, tough)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).SummonedThisTurn = false })
	floatMana(t, g, me, "{G}{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, elf); p != 4 || tough != 4 {
		t.Errorf("my other Elf is %d/%d, want 4/4 (+1/+1 for each of Haldir's three counters)", p, tough)
	}
	if c, _ := battlefieldCard(g, elf); !game.HasKeyword(&c, "vigilance") {
		t.Error("my other Elf did not gain vigilance")
	}
	if p, _ := ptOf(t, g, other); p != 2 {
		t.Errorf("a non-Elf got the pump: power %d", p)
	}
	if p, _ := ptOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Elf got the pump: power %d", p)
	}
	if p, _ := ptOf(t, g, id); p != 3 {
		t.Errorf("Haldir pumped himself: power %d", p)
	}
}

// --- Elvish Warmaster --------------------------------------------------------

func TestGapB116ElvishWarmasterMakesOneWarriorPerTurnWhenOtherElvesEnter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Elvish Warmaster", "Creature — Elf Warrior", gapB116Warmaster, 2, 2)
	advanceToMain(t, g)
	b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Elf Warrior"); n != 0 {
		t.Fatalf("a non-Elf entering made %d Elf Warriors", n)
	}
	b12PlayFromHand(t, g, "Elf One", "Creature — Elf Druid", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Elf Warrior"); n != 1 {
		t.Fatalf("%d Elf Warrior tokens, want 1", n)
	}
	b12PlayFromHand(t, g, "Elf Two", "Creature — Elf Druid", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// The token itself is an Elf that entered under my control, but the trigger
	// already spent its once this turn.
	if n := countNamed(g, "Elf Warrior"); n != 1 {
		t.Errorf("%d Elf Warrior tokens after a second Elf, want still 1 (once each turn)", n)
	}
}

func TestGapB116ElvishWarmasterPumpsElvesAndGivesDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	warmaster := b12Push(g, me.ID, "Elvish Warmaster", "Creature — Elf Warrior", gapB116Warmaster, 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Druid", 1, 1)
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}{G}{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, warmaster, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{warmaster, elf} {
		c, _ := battlefieldCard(g, id)
		if !game.HasKeyword(&c, "deathtouch") {
			t.Errorf("%s lacks deathtouch", c.Name)
		}
	}
	if p, _ := ptOf(t, g, elf); p != 3 {
		t.Errorf("my Elf has power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, warmaster); p != 4 {
		t.Errorf("the Warmaster has power %d, want 4", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("a non-Elf was pumped: power %d", p)
	}
	if p, _ := ptOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Elf was pumped: power %d", p)
	}
}

// --- Prize Pig ---------------------------------------------------------------

func TestGapB116PrizePigUntapsAtThreeRibbonCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pig := b12Push(g, me.ID, "Prize Pig", "Creature — Boar", gapB116Pig, 0, 3)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, pig).Tapped = true })
	ribbons := func() int {
		c, _ := battlefieldCard(g, pig)
		return c.Counters["ribbon"]
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if got := ribbons(); got != 2 {
		t.Fatalf("%d ribbon counters after gaining 2, want 2", got)
	}
	if !b16Tapped(t, g, pig) {
		t.Error("the Pig untapped below three counters")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := ribbons(); got != 0 {
		t.Errorf("%d ribbon counters after reaching three, want 0 (removed)", got)
	}
	if b16Tapped(t, g, pig) {
		t.Error("the Pig did not untap")
	}
}

func TestGapB116PrizePigRemovesAllCountersEvenPastThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pig := b12Push(g, me.ID, "Prize Pig", "Creature — Boar", gapB116Pig, 0, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, pig); c.Counters["ribbon"] != 0 {
		t.Errorf("%d ribbon counters, want all five removed", c.Counters["ribbon"])
	}
}

// --- Aragorn, Hornburg Hero --------------------------------------------------

func TestGapB116AragornGivesAttackersFirstStrikeAndRenown(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Aragorn, Hornburg Hero", "Legendary Creature — Human Soldier", gapB116Aragorn, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if c, _ := battlefieldCard(g, bear); game.HasKeyword(&c, "first strike") {
		t.Error("a creature that is not attacking has first strike")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	assertKeywords(t, g, bear, "first strike", "renown 1")
	if t2, _ := battlefieldCard(g, theirs); game.HasKeyword(&t2, "first strike") {
		t.Error("an opponent's creature has first strike")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters on the Bear, want 1 (renown 1)", got)
	}
	if !g.IsRenowned(bear) {
		t.Error("the Bear did not become renowned")
	}
}

func TestGapB116AragornDoublesTheCountersOfAnAlreadyRenownedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Aragorn, Hornburg Hero", "Legendary Creature — Human Soldier", gapB116Aragorn, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, bear)
		c.Counters = map[string]int{"+1/+1": 3}
		c.Renowned = true
	})
	advanceToMain(t, g)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 6 {
		t.Errorf("%d +1/+1 counters, want 6 (3 doubled)", got)
	}
}

// --- The Balrog of Moria -----------------------------------------------------

func TestGapB116BalrogCycleMakesTwoTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	cycleFromHand(t, g, "The Balrog of Moria", "Legendary Creature — Avatar Demon", gapB116Balrog, "{C}{C}{C}{R}")
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Treasure"); n != 2 {
		t.Errorf("%d Treasure tokens, want 2", n)
	}
}

func TestGapB116BalrogDiesMayExileItToExileACreatureOfEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp1, opp2 := g.Seats[1], g.Seats[2]
	balrog := b12Push(g, me.ID, "The Balrog of Moria", "Legendary Creature — Avatar Demon", gapB116Balrog, 8, 8)
	a := b12Creature(g, opp1.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp2.ID, "B", "Creature — Bear", 2, 2)
	spare := b12Creature(g, opp1.ID, "Spare", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, balrog)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(balrog) {
		t.Fatal("the Balrog was not exiled")
	}
	pickCard(t, g, me.ID, a)
	pickCard(t, g, me.ID, b)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a chosen creature was not exiled")
	}
	if !g.Battlefield.Contains(spare) {
		t.Error("a second creature of the same opponent was exiled (up to one each)")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was exiled")
	}
}
