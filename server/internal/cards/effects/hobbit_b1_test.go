package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB) cards, half B, batch 1. Every test drives the
// registered catalog row through the engine.

func hobOracle(t *testing.T, name string) string {
	t.Helper()
	for _, s := range All() {
		if s.Name == name {
			return s.OracleID
		}
	}
	t.Fatalf("%s is not registered", name)
	return ""
}

// hobCast casts the named creature from the active seat's hand and lets
// it and its triggers settle.
func hobCast(t *testing.T, g *game.Game, name, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	return rfdCast(t, g, name, typeLine, hobOracle(t, name), power, toughness)
}

// hobPush puts the named catalog creature onto the battlefield, ready to
// attack.
func hobPush(t *testing.T, g *game.Game, who uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	return b12Push(g, who, name, typeLine, hobOracle(t, name), power, toughness)
}

func hobArmyCounters(t *testing.T, g *game.Game, who uuid.UUID) int {
	t.Helper()
	return armyOf(t, g, who).Counters[game.CounterPlusOne]
}

func hobDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
}

func TestHobStonyVoicedGoblinsEachOpponentDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	hobCast(t, g, "Stony-Voiced Goblins", "Creature — Goblin Bard", 1, 1)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("the caster must not be asked to discard")
	}
	for _, p := range g.Seats[1:] {
		if discardChoiceFor(g, p.ID) == nil {
			t.Fatalf("opponent %s was not asked to discard", p.ID)
		}
		discardFromHand(t, g, p.ID)
		if got := p.Hand.Size(); got != before[p.ID]-1 {
			t.Errorf("opponent hand = %d, want %d", got, before[p.ID]-1)
		}
	}
}

func TestHobLargeBearHasReachTrampleAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	id := hobPush(t, g, g.Seats[0].ID, "Large Bear", "Creature — Bear", 5, 5)
	for _, kw := range []string{"reach", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Large Bear lacks %s", kw)
		}
	}
	if hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Large Bear must not fly")
	}
}

func TestHobGiganticBigBearCantBeCounteredAndIsHexproofHasty(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Gigantic Big Bear", "Creature — Bear", hobOracle(t, "Gigantic Big Bear"), nil)
	var uncounterable bool
	g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("Gigantic Big Bear on the stack can be countered")
	}
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, id) {
		t.Fatal("Gigantic Big Bear did not resolve")
	}
	for _, kw := range []string{"hexproof", "haste"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Gigantic Big Bear lacks %s", kw)
		}
	}
}

func TestHobFearsomeGoblinPairAmassesFourWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pair := hobPush(t, g, me.ID, "Fearsome Goblin Pair", "Creature — Goblin Soldier", 1, 1)
	hobDestroy(t, g, pair)
	army := armyOf(t, g, me.ID)
	if army.Counters[game.CounterPlusOne] != 4 || !army.HasSubtype("Goblin") {
		t.Fatalf("Army = %+v, want a Goblin Army with four +1/+1 counters", army)
	}
	// A second one adds to the same Army rather than making another.
	pair2 := hobPush(t, g, me.ID, "Fearsome Goblin Pair", "Creature — Goblin Soldier", 1, 1)
	hobDestroy(t, g, pair2)
	if got := hobArmyCounters(t, g, me.ID); got != 8 {
		t.Fatalf("Army counters = %d, want 8", got)
	}
}

func TestHobFearsomeGoblinPairDoesNothingWhileAlive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "Fearsome Goblin Pair", "Creature — Goblin Soldier", 1, 1)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if game.IsArmy(c) {
			t.Fatal("an Army appeared before the Pair died")
		}
	}
}

func TestHobGoblinTownFlunkiesHasHasteAndAmassesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobCast(t, g, "Goblin-town Flunkies", "Creature — Goblin Soldier", 1, 1)
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("Goblin-town Flunkies lacks haste")
	}
	if got := hobArmyCounters(t, g, me.ID); got != 1 {
		t.Fatalf("Army counters = %d, want 1", got)
	}
	hobCast(t, g, "Goblin-town Flunkies", "Creature — Goblin Soldier", 1, 1)
	if got := hobArmyCounters(t, g, me.ID); got != 2 {
		t.Fatalf("a second Flunkies: Army counters = %d, want 2 on the same Army", got)
	}
}

func TestHobBothersomeNoisemakerAmassesOnNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "Bothersome Noisemaker", "Creature — Goblin Bard", 2, 2)
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if got := hobArmyCounters(t, g, me.ID); got != 1 {
		t.Fatalf("Army counters after a noncreature spell = %d, want 1", got)
	}
	castCatalogSpell(t, g, "Plain Creature", "Creature — Test", "", nil)
	passPriorityAroundTable(t, g)
	if got := hobArmyCounters(t, g, me.ID); got != 1 {
		t.Fatalf("Army counters after a creature spell = %d, want still 1", got)
	}
}

// attackFerocious attacks with `attacker`, with or without a creature of
// power 4 on the attacking player's side.
func hobAttackSetup(t *testing.T, big bool, name, typeLine string, power, toughness int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, name, typeLine, power, toughness)
	if big {
		pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
	}
	attackWith(t, g, g.Seats[1].ID, id)
	return g, id
}

func TestHobNighthowlPursuerFerociousPump(t *testing.T) {
	for _, big := range []bool{true, false} {
		g, id := hobAttackSetup(t, big, "Nighthowl Pursuer", "Creature — Wolf", 1, 1)
		c := gaLayered(t, g, id)
		want := 1
		if big {
			want = 3
		}
		if c.CurrentPower() != want || c.CurrentToughness() != want {
			t.Errorf("big ally %v: Pursuer is %d/%d, want %d/%d", big, c.CurrentPower(), c.CurrentToughness(), want, want)
		}
		if !hasEffectiveKeyword(t, g, id, "menace") {
			t.Error("Nighthowl Pursuer lacks menace")
		}
	}
}

func TestHobNighthowlPursuerCountsItselfOnceItIsBigEnough(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, "Nighthowl Pursuer", "Creature — Wolf", 1, 1)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, 3); err != nil {
			t.Fatal(err)
		}
	})
	attackWith(t, g, g.Seats[1].ID, id)
	if got := gaLayered(t, g, id).CurrentPower(); got != 6 {
		t.Fatalf("a 4/4 Pursuer attacking alone has power %d, want 6", got)
	}
}

func TestHobRaveningWargFerociousLifeGain(t *testing.T) {
	for _, big := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := hobPush(t, g, me.ID, "Ravening Warg", "Creature — Wolf", 2, 2)
		if big {
			pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
		}
		if !hasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("Ravening Warg lacks deathtouch")
		}
		life := me.Life
		attackWith(t, g, g.Seats[1].ID, id)
		want := life
		if big {
			want = life + 2
		}
		if me.Life != want {
			t.Errorf("big ally %v: life %d, want %d", big, me.Life, want)
		}
	}
}

func TestHobWilderlandScroungerFerociousCountersOnEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, "Wilderland Scrounger", "Creature — Wolf", 3, 6)
	big := pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
	other := pushVanillaCreature(g, me.ID, "Small Ally", 1, 1)
	theirs := pushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	attackWith(t, g, g.Seats[1].ID, id)
	for _, cid := range []uuid.UUID{id, big, other} {
		if got := gaLayered(t, g, cid).Counters[game.CounterPlusOne]; got != 1 {
			t.Errorf("creature %s has %d +1/+1 counters, want 1", cid, got)
		}
	}
	if got := gaLayered(t, g, theirs).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestHobWilderlandScroungerNeedsAFourPowerCreature(t *testing.T) {
	g, id := hobAttackSetup(t, false, "Wilderland Scrounger", "Creature — Wolf", 3, 6)
	if got := gaLayered(t, g, id).Counters[game.CounterPlusOne]; got != 0 {
		t.Fatalf("Scrounger got %d counters without a power-4 creature", got)
	}
}

func TestHobNastyLittleRabbitGrowsAtBeginningOfCombat(t *testing.T) {
	for _, big := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := hobPush(t, g, me.ID, "Nasty Little Rabbit", "Creature — Rabbit", 1, 2)
		if big {
			pushVanillaCreature(g, me.ID, "Big Ally", 4, 4)
		}
		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		want := 0
		if big {
			want = 1
		}
		if got := gaLayered(t, g, id).Counters[game.CounterPlusOne]; got != want {
			t.Errorf("big ally %v: counters %d, want %d", big, got, want)
		}
	}
}

func TestHobNastyLittleRabbitIgnoresAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := hobPush(t, g, opp.ID, "Nasty Little Rabbit", "Creature — Rabbit", 1, 2)
	pushVanillaCreature(g, opp.ID, "Big Ally", 4, 4)
	advanceTo(t, g, game.StepBeginCombat) // seat 0's combat
	passPriorityAroundTable(t, g)
	if got := gaLayered(t, g, id).Counters[game.CounterPlusOne]; got != 0 {
		t.Fatalf("Rabbit grew %d on someone else's turn", got)
	}
}
