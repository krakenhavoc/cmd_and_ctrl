package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_test2_test.go — Queza and Squall, slice raffine-a (#3212).

const (
	quezaOracle  = "c983338d-ae6b-4a15-9931-daa20fca8269"
	squallOracle = "e2348e0e-eeca-47b5-9417-bf4470ec29bb"
)

func TestQuezaDrainsTheChosenOpponentPerCardDrawn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Queza, Augur of Agonies", "Legendary Creature — Octopus Advisor", quezaOracle, false)
	myLife, oppLife := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) {
		t.Error("Queza offered its own controller as the opponent")
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("life me %d→%d opp %d→%d, want +1 / -1", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestQuezaIgnoresAnOpponentsDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Queza, Augur of Agonies", "Legendary Creature — Octopus Advisor", quezaOracle, false)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if latestPickTarget(g, me.ID) != nil {
		t.Error("Queza triggered on an opponent's draw")
	}
}

// Rough Divide: one attacker gains double strike; two attackers gain none.
func TestSquallGivesAnAloneAttackerDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Squall, SeeD Mercenary", "Legendary Creature — Human Knight Mercenary", squallOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, a, "double strike") {
		t.Error("the lone attacker did not gain double strike")
	}
}

func TestSquallGivesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Squall, SeeD Mercenary", "Legendary Creature — Human Knight Mercenary", squallOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if effectiveAbilitiesContain(t, g, a, "double strike") || effectiveAbilitiesContain(t, g, b, "double strike") {
		t.Error("a creature gained double strike though two attacked")
	}
}

// Squall connecting returns a cheap permanent card, and only a cheap one.
func TestSquallCombatDamageReturnsACheapPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squall := pushDiesCreatureForTest(g, me.ID, "Squall, SeeD Mercenary", squallOracle, "Legendary Creature — Human Knight Mercenary", 3, 4)
	cheap := pushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	dear := pushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")
	theirs := pushGraveyardPermanent(opp, "Their Bears", "Creature — Bear", "{1}{G}")

	// Squall attacks alone, so Rough Divide gives it double strike and the
	// first-strike damage step is where it first connects.
	declareAttack(t, g, opp.ID, squall)
	passPriorityAroundTable(t, g)
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt after combat damage")
	}
	if !hasID(prompt.PickTargetCards, cheap) {
		t.Error("the mana-value-2 permanent card is not offered")
	}
	if hasID(prompt.PickTargetCards, dear) || hasID(prompt.PickTargetCards, theirs) {
		t.Error("a mana-value-5 card or an opponent's card was offered")
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the chosen card did not return to the battlefield")
	}
	if opp.Life != 40-3 {
		t.Errorf("opponent life %d, want 37 after the first strike", opp.Life)
	}
}

// Blocked, Squall deals no damage to a player, so nothing returns.
func TestSquallBlockedReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squall := pushDiesCreatureForTest(g, me.ID, "Squall, SeeD Mercenary", squallOracle, "Legendary Creature — Human Knight Mercenary", 3, 4)
	cheap := pushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 9)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(squall, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, squall); err != nil {
		t.Fatalf("block: %v", err)
	}
	for i := 0; i < 4 && g.Turn.Step != game.StepEndCombat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	if latestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(cheap) {
		t.Error("Squall returned a card though it was blocked")
	}
}
