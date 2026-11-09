package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b9_cards_test.go — Final Fantasy batches 9 and 10 (#4462).

const (
	finB9GabranthOracle = "1a846ad6-15c9-4bf9-9127-612018a557b9"
)

var _ = uuid.Nil

func TestFinB9GabranthGrowsWhenYourOtherCreatureOrArtifactDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gab := b12Push(g, me.ID, "Judge Magister Gabranth", "Legendary Creature — Human Advisor Knight", finB9GabranthOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	art := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if !hasEffectiveKeyword(t, g, gab, "menace") {
		t.Error("no menace")
	}
	for i, id := range []uuid.UUID{bear, art, theirs, land} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
		passPriorityAroundTable(t, g)
		want := []int{1, 2, 2, 2}[i]
		if got := countersOn(g, gab, game.CounterPlusOne); got != want {
			t.Fatalf("after death %d: %d counters, want %d", i, got, want)
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gab) })
	passPriorityAroundTable(t, g)
}

const finB9JenovaOracle = "9370f6bf-d6a0-4c5b-828f-7229f563cc07"

func TestFinB9JenovaCountersMutatesAndDrawsOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jen := b12Push(g, me.ID, "Jenova, Ancient Calamity", "Legendary Creature — Alien", finB9JenovaOracle, 1, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("bear has %d counters, want 1 (Jenova's power)", got)
	}
	if c, _ := g.LookupCardForEffect(bear); !c.HasSubtype("Mutant") || !c.HasSubtype("Bear") {
		t.Error("bear is not a Bear and a Mutant")
	}
	_ = jen
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3 (a 3-power Mutant)", hand, me.Hand.Size())
	}
}

const (
	finB9BoltOracle  = "aad9f9f2-4c6b-48d0-80bd-77b73a9ea5f1"
	finB9RufusOracle = "0caeb396-e574-4e80-a1f6-79988fe9ba19"
	finB9RinoaOracle = "8896c74b-d466-4c25-94cf-82b5e2b1fdef"
)

func TestFinB9JudgmentBoltHitsTheCreatureAndItsControllerPerEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	finABEquipment(g, me.ID, "Sword")
	finABEquipment(g, me.ID, "Shield")
	finABEquipment(g, opp.ID, "Their Sword")
	life := opp.Life
	finABCastInstantNow(t, g, "Judgment Bolt", finB9BoltOracle, finABCard(victim))
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("their life %d → %d, want -2 (two Equipment of yours)", life, opp.Life)
	}
	if !opp.Graveyard.Contains(victim) {
		t.Error("the 5-damage creature survived")
	}
}

func TestFinB9JudgmentBoltWithNoEquipmentDealsOnlyTheFive(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	life := opp.Life
	finABCastInstantNow(t, g, "Judgment Bolt", finB9BoltOracle, finABCard(victim))
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("their life %d → %d, want unchanged", life, opp.Life)
	}
	if opp.Graveyard.Contains(victim) {
		t.Error("a 6-toughness creature died to 5")
	}
}

func TestFinB9RufusMakesDarkstarOnlyWithoutOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rufus := b12Push(g, me.ID, "Rufus Shinra", "Legendary Creature — Human Noble", finB9RufusOracle, 2, 4)
	finABAttack(t, g, opp.ID, rufus)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Darkstar"); n != 1 {
		t.Fatalf("%d Darkstar tokens after the first attack, want 1", n)
	}
}

func TestFinB9RufusDoesNotMakeASecondDarkstar(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rufus := b12Push(g, me.ID, "Rufus Shinra", "Legendary Creature — Human Noble", finB9RufusOracle, 2, 4)
	b12Creature(g, me.ID, "Darkstar", "Legendary Creature — Dog", 2, 2)
	finABAttack(t, g, opp.ID, rufus)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Darkstar"); n != 1 {
		t.Errorf("%d Darkstars, want the one you already control", n)
	}
}

func TestFinB9RinoaMakesAngeloAndPumpsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rinoa := b12Push(g, me.ID, "Rinoa Heartilly", "Legendary Creature — Human Rebel Warlock", finB9RinoaOracle, 4, 4)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	finABAttack(t, g, opp.ID, rinoa)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(ally)...)
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, ally)
	if v.CurrentPower() != 5 || v.CurrentToughness() != 5 {
		t.Errorf("ally is %d/%d, want 5/5 (+3 for three creatures)", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestFinB9RinoaEntersWithAngelo(t *testing.T) {
	g := newCatalogGame(t)
	castCreatureCard(t, g, "Rinoa Heartilly", "Legendary Creature — Human Rebel Warlock", finB9RinoaOracle, 4, 4)
	if n := onBattlefieldNamed(g, "Angelo"); n != 1 {
		t.Errorf("%d Angelo tokens, want 1", n)
	}
}
