package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_test2_test.go — Queza and Squall, slice raffine-a (#3212).

const (
	quezaOracle  = "c983338d-ae6b-4a15-9931-daa20fca8269"
	squallOracle = "e2348e0e-eeca-47b5-9417-bf4470ec29bb"
)

func TestQuezaDrainsTheChosenOpponentPerCardDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Queza, Augur of Agonies", "Legendary Creature — Octopus Advisor", quezaOracle, false)
	myLife, oppLife := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("Queza offered its own controller as the opponent")
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("life me %d→%d opp %d→%d, want +1 / -1", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestQuezaIgnoresAnOpponentsDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Queza, Augur of Agonies", "Legendary Creature — Octopus Advisor", quezaOracle, false)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("Queza triggered on an opponent's draw")
	}
}

// Rough Divide: one attacker gains double strike; two attackers gain none.
func TestSquallGivesAnAloneAttackerDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Squall, SeeD Mercenary", "Legendary Creature — Human Knight Mercenary", squallOracle, false)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, a, "double strike") {
		t.Error("the lone attacker did not gain double strike")
	}
}

func TestSquallGivesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Squall, SeeD Mercenary", "Legendary Creature — Human Knight Mercenary", squallOracle, false)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, a, "double strike") || effectstest.EffectiveAbilitiesContain(t, g, b, "double strike") {
		t.Error("a creature gained double strike though two attacked")
	}
}

// Squall connecting returns a cheap permanent card, and only a cheap one.
func TestSquallCombatDamageReturnsACheapPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squall := effectstest.PushDiesCreatureForTest(g, me.ID, "Squall, SeeD Mercenary", squallOracle, "Legendary Creature — Human Knight Mercenary", 3, 4)
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	dear := effectstest.PushGraveyardPermanent(me, "Serra Angel", "Creature — Angel", "{3}{W}{W}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Bears", "Creature — Bear", "{1}{G}")

	// Squall attacks alone, so Rough Divide gives it double strike and the
	// first-strike damage step is where it first connects.
	effectstest.DeclareAttack(t, g, opp.ID, squall)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt after combat damage")
	}
	if !effectstest.HasID(prompt.PickTargetCards, cheap) {
		t.Error("the mana-value-2 permanent card is not offered")
	}
	if effectstest.HasID(prompt.PickTargetCards, dear) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Error("a mana-value-5 card or an opponent's card was offered")
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the chosen card did not return to the battlefield")
	}
	if opp.Life != 40-3 {
		t.Errorf("opponent life %d, want 37 after the first strike", opp.Life)
	}
}

// Blocked, Squall deals no damage to a player, so nothing returns.
func TestSquallBlockedReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squall := effectstest.PushDiesCreatureForTest(g, me.ID, "Squall, SeeD Mercenary", squallOracle, "Legendary Creature — Human Knight Mercenary", 3, 4)
	cheap := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(squall, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, squall); err != nil {
		t.Fatalf("block: %v", err)
	}
	for i := 0; i < 4 && g.Turn.Step != game.StepEndCombat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(cheap) {
		t.Error("Squall returned a card though it was blocked")
	}
}
