package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o336Arcbond   = "01790829-cc03-430f-a899-2f274b7dbf62"
	o336Contest   = "3b78f940-f840-4d4d-a09b-147202963367"
	o336Keg       = "8e819d68-3a6e-449a-8392-e072724b80f1"
	o336Mindscour = "c5d326a9-c047-4f77-9f0c-b30daf268f15"
	o336Friendly  = "d4ab3e1d-e1c4-43c3-8068-1809938ab094"
	o336Shockmaw  = "e0754b8e-fd3c-4207-9031-a6323e1b66d5"
	o336Harsh     = "f5cf471f-032c-4ebb-823c-8841de209216"
)

func TestO336CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o336Arcbond: "Arcbond", o336Contest: "Grim Contest", o336Keg: "Goblin Boom Keg",
		o336Mindscour: "Mindscour Dragon", o336Friendly: "Friendly Fire", o336Shockmaw: "Shockmaw Dragon",
		o336Harsh: "Harsh Sustenance",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o336Card(id uuid.UUID) []game.TargetRef { return effectstest.CardRefs(id) }

func TestO336ArcbondReflectsEachDamageToEveryOtherCreatureAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chosen := effectstest.B12Creature(g, me.ID, "Chosen", "Creature — Giant", 0, 9)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 2)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Bear", 2, 6)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Arcbond", "Instant", o336Arcbond, o336Card(chosen))
	effectstest.PassPriorityAroundTable(t, g)
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.DealDamageTo(t, g, small, chosen, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(other) {
		t.Error("a 2-damage reflection should kill the 1/2 and the 1/1")
	}
	if !g.Battlefield.Contains(sturdy) || !g.Battlefield.Contains(chosen) {
		t.Error("the sturdy creature or the chosen creature died")
	}
	if me.Life != lifeMe-2 || opp.Life != lifeOpp-2 {
		t.Errorf("life %d/%d, want each player down 2 (%d/%d)", me.Life, opp.Life, lifeMe-2, lifeOpp-2)
	}
	// It triggers again for every later damage this turn.
	effectstest.DealDamageTo(t, g, sturdy, chosen, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-3 {
		t.Errorf("opponent life %d, want %d after a second hit", opp.Life, lifeOpp-3)
	}
}

func TestO336ArcbondIgnoresDamageToAnyOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chosen := effectstest.B12Creature(g, me.ID, "Chosen", "Creature — Giant", 0, 9)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 1, 5)
	effectstest.CastCatalogSpell(t, g, "Arcbond", "Instant", o336Arcbond, o336Card(chosen))
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.DealDamageTo(t, g, chosen, other, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Error("damage to another creature triggered the reflection")
	}
}

func TestO336GrimContestEachDealsItsToughnessToTheOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 9, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Grim Contest", "Instant", o336Contest, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 1},
	}); err == nil {
		t.Fatal("Grim Contest took the opponent's creature as its own")
	}
	effectstest.CastCatalogSpell(t, g, "Grim Contest", "Instant", o336Contest, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("my 0/3 took 4 damage and survived")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("their 4-toughness creature took only 3 damage and died")
	}
}

func TestO336KegSacrificesAtUpkeepThenDealsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keg := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Boom Keg", "Artifact", o336Keg, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(keg) {
		t.Fatal("the Keg was not sacrificed at the beginning of my upkeep")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the Keg's 3 damage did not kill the 2/3")
	}
}

func TestO336KegPaysOffWhenDestroyedToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keg := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Boom Keg", "Artifact", o336Keg, false)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(keg) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
}

func TestO336MindscourMillsTheChosenPlayerOnCombatDamageToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Mindscour Dragon", "Creature — Dragon", o336Mindscour, 4, 4)
	lib := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 4 {
		t.Errorf("my library lost %d cards, want 4 (the target is any player)", got)
	}
}

func TestO336ShockmawDamagesEachCreatureOfThePlayerItHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Shockmaw Dragon", "Creature — Dragon", o336Shockmaw, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	weak := effectstest.B12Creature(g, opp.ID, "Weak", "Creature — Bear", 1, 1)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Bear", 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(weak) {
		t.Error("the 1/1 took 1 damage and survived")
	}
	if !g.Battlefield.Contains(sturdy) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that should not have died did")
	}
}

func TestO336FriendlyFireHitsTheCreatureAndItsControllerForTheRevealedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ox", 2, 5)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	effectstest.PutInHand(opp, game.Card{Name: "Expensive", TypeLine: "Sorcery", ManaCost: "{2}{R}"})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Friendly Fire", "Instant", o336Friendly, o336Card(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the controller lost %d life, want 3", life-opp.Life)
	}
	if got := effectstest.DamageOn(t, g, victim); got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
}

func TestO336FriendlyFireWithAnEmptyHandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ox", 2, 5)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Friendly Fire", "Instant", o336Friendly, o336Card(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || effectstest.DamageOn(t, g, victim) != 0 {
		t.Error("damage was dealt with nothing to reveal")
	}
}

func TestO336HarshSustenanceDealsAndGainsTheNumberOfCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Harsh Sustenance", "Instant", o336Harsh, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-2 || me.Life != lifeMe+2 {
		t.Errorf("life me %d opp %d, want %d and %d", me.Life, opp.Life, lifeMe+2, lifeOpp-2)
	}
}
