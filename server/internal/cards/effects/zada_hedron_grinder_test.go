package effects

import (
	"testing"

	"github.com/google/uuid"
)

const zadaHedronGrinderOracle = "9f7d5e1c-bf5b-4ab9-bd6e-2627b4d22b36"

func TestZadaCopiesTheSpellForEachOtherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	zada := b12Push(g, me.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(zada))
	if n := triggersOnStackFrom(g, zada); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no order prompt for the controller")
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, a) || !hasID(pick.ChooseCards, b) {
		t.Fatalf("order prompt over %v, want the two other creatures I control", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, b, a)
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{a, b} {
		if p := effectivePower(t, g, id); p != 5 {
			t.Errorf("my creature is %d power, want 5", p)
		}
	}
	if p := effectivePower(t, g, zada); p != 6 {
		t.Errorf("Zada is %d power, want 6", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
}

func TestZadaIgnoresASpellWithAnotherTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	zada := b12Push(g, me.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(a))
	if n := triggersOnStackFrom(g, zada); n != 0 {
		t.Errorf("a spell targeting another creature triggered Zada %d times", n)
	}
}

func TestZadaDoesNotTriggerOnAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	zada := b12Push(g, opp.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(zada))
	if n := triggersOnStackFrom(g, zada); n != 0 {
		t.Errorf("an opponent's spell triggered Zada %d times (it is \"you cast\")", n)
	}
}

func TestZadaSkipsACreatureTheSpellCantTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	zada := b12Push(g, me.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	shy := b12Creature(g, me.ID, "Shy Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shy {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "shroud")
			}
		}
	})
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(zada))
	passPriorityAroundTable(t, g)
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatal("an order prompt with one creature the spell could target")
	}
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	if p := effectivePower(t, g, a); p != 5 {
		t.Errorf("Bear A is %d power, want 5", p)
	}
	if p := effectivePower(t, g, shy); p != 2 {
		t.Errorf("the shroud Bear is %d power, want 2", p)
	}
}
