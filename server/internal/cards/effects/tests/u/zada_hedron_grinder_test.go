package testsu

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const zadaHedronGrinderOracle = "9f7d5e1c-bf5b-4ab9-bd6e-2627b4d22b36"

func TestZadaCopiesTheSpellForEachOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	zada := effectstest.B12Push(g, me.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(zada))
	if n := effectstest.TriggersOnStackFrom(g, zada); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no order prompt for the controller")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, a) || !effectstest.HasID(pick.ChooseCards, b) {
		t.Fatalf("order prompt over %v, want the two other creatures I control", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b, a)
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{a, b} {
		if p := effectstest.EffectivePower(t, g, id); p != 5 {
			t.Errorf("my creature is %d power, want 5", p)
		}
	}
	if p := effectstest.EffectivePower(t, g, zada); p != 6 {
		t.Errorf("Zada is %d power, want 6", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
}

func TestZadaIgnoresASpellWithAnotherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zada := effectstest.B12Push(g, me.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(a))
	if n := effectstest.TriggersOnStackFrom(g, zada); n != 0 {
		t.Errorf("a spell targeting another creature triggered Zada %d times", n)
	}
}

func TestZadaDoesNotTriggerOnAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	zada := effectstest.B12Push(g, opp.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(zada))
	if n := effectstest.TriggersOnStackFrom(g, zada); n != 0 {
		t.Errorf("an opponent's spell triggered Zada %d times (it is \"you cast\")", n)
	}
}

func TestZadaSkipsACreatureTheSpellCantTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zada := effectstest.B12Push(g, me.ID, "Zada, Hedron Grinder", "Legendary Creature — Goblin Ally", zadaHedronGrinderOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	shy := effectstest.B12Creature(g, me.ID, "Shy Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == shy {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "shroud")
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.SocB8GiantGrowthOracle, effectstest.CardTarget(zada))
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatal("an order prompt with one creature the spell could target")
	}
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p := effectstest.EffectivePower(t, g, a); p != 5 {
		t.Errorf("Bear A is %d power, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, shy); p != 2 {
		t.Errorf("the shroud Bear is %d power, want 2", p)
	}
}
