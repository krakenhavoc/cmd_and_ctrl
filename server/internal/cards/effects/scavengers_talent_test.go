package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const scavengersTalentOracle = "aefea339-8a0d-4531-8a62-afaecc88d078"

func stPushTalent(g *game.Game, owner uuid.UUID, level int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Scavenger's Talent", TypeLine: "Enchantment — Class", OracleID: scavengersTalentOracle,
		Owner: owner, Controller: owner, ClassLevel: level,
	})
}

func TestScavengersTalentIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(scavengersTalentOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Scavenger's Talent: %+v %v", spec, ok)
	}
}

func TestScavengersTalentLevelOneMakesOneFoodPerTurnForADeathBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stPushTalent(g, me.ID, 1)
	a := pushCatalogPermanent(g, me.ID, "A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, me.ID, "B", "Creature — Bear", "", false)
	c := pushCatalogPermanent(g, me.ID, "C", "Creature — Bear", "", false)
	killCreature(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Food"); n != 1 {
		t.Fatalf("Food after one death: %d, want 1", n)
	}
	killCreature(t, g, me.ID, b)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Food"); n != 1 {
		t.Errorf("Food after a second death the same turn: %d, want still 1", n)
	}
	_ = c
}

func TestScavengersTalentLevelOneIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stPushTalent(g, me.ID, 1)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	killCreature(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Food"); n != 0 {
		t.Errorf("Food: %d, want 0", n)
	}
}

func TestScavengersTalentLevelTwoMillsTwoPerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stPushTalent(g, me.ID, 2)
	seedLibrary(opp, "1", "2", "3", "4", "5")
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	size := opp.Library.Size()
	sacrificeNow(t, g, me.ID, rock)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := size - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestScavengersTalentLevelTwoIsOffAtLevelOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stPushTalent(g, me.ID, 1)
	seedLibrary(opp, "1", "2", "3")
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	size := opp.Library.Size()
	sacrificeNow(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != size {
		t.Error("the level-2 ability worked at level 1")
	}
}

func TestScavengersTalentLevelThreeSacrificesThreeAndReturnsWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	talent := stPushTalent(g, me.ID, 3)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Old Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	fodder := []uuid.UUID{
		pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false),
		pushCatalogPermanent(g, me.ID, "Ring", "Enchantment", "", false),
		pushCatalogPermanent(g, me.ID, "Cub", "Creature — Bear", "", false),
	}
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	pay := stPayUnlessFor(g, me.ID)
	if pay == nil {
		t.Fatal("no may-sacrifice offer at the end step")
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.PayActionOptionsForEffect(me.ID, pay.PayAction()) })
	for _, bad := range []uuid.UUID{talent, land} {
		for _, o := range opts {
			if o == bad {
				t.Errorf("the offer includes the Talent itself or a land")
			}
		}
	}
	if err := g.ResolvePayUnlessWithCards(pay.ID, me.ID, true, fodder); err != nil {
		t.Fatalf("pay: %v", err)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature card to return was offered")
	}
	// The sacrificed creature card is among the candidates.
	if len(pick.ChooseCards) != 2 {
		t.Errorf("candidates %d, want the dead Bear and the sacrificed Cub", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	c, ok := battlefieldCard(g, dead)
	if !ok || counterOn(g, dead, game.CounterFinality) != 1 {
		t.Fatalf("returned=%v finality=%d, want a Bear back with a finality counter", ok, counterOn(g, dead, game.CounterFinality))
	}
	_ = c
}

func TestScavengersTalentLevelThreeNeedsThreeOtherNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stPushTalent(g, me.ID, 3)
	pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	pushCatalogPermanent(g, me.ID, "Ring", "Enchantment", "", false)
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if stPayUnlessFor(g, me.ID) != nil {
		t.Error("offered a sacrifice with only two other nonland permanents")
	}
}

func stPayUnlessFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
