package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const garrukCursedHuntsmanOracle = "e0cef79c-ad47-4cbc-9d73-a913e487ccb7"

func TestGarrukCursedHuntsmanWolvesFeedEachGarrukWhenTheyDie(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWalkerForTest(g, me.ID, "Garruk, Cursed Huntsman", garrukCursedHuntsmanOracle, 5)
	other := pushWalkerForTest(g, me.ID, "Garruk Test Twin", "", 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == other || g.Battlefield.Cards[i].InstanceID == pw {
				g.Battlefield.Cards[i].TypeLine = "Legendary Planeswalker — Garruk"
			}
		}
	})
	enemyWalker := pushWalkerForTest(g, g.Seats[(seat+1)%len(g.Seats)].ID, "Enemy Garruk", "", 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == enemyWalker {
				g.Battlefield.Cards[i].TypeLine = "Legendary Planeswalker — Garruk"
			}
		}
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Wolf"); n != 2 {
		t.Fatalf("wolves = %d, want 2", n)
	}
	if got := loyaltyOf(g, pw); got != 5 {
		t.Fatalf("loyalty after 0 = %d, want 5", got)
	}
	var wolf uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wolf" {
			wolf = c.InstanceID
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wolf, game.DestroyOptions{}) })
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, pw); got != 6 {
		t.Errorf("Garruk loyalty = %d, want 6", got)
	}
	if got := loyaltyOf(g, other); got != 4 {
		t.Errorf("second Garruk loyalty = %d, want 4 (each Garruk you control)", got)
	}
	if got := loyaltyOf(g, enemyWalker); got != 3 {
		t.Errorf("an opponent's Garruk gained loyalty: %d", got)
	}
}

func TestGarrukCursedHuntsmanMinusThreeDestroysAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, me.ID, "Garruk, Cursed Huntsman", garrukCursedHuntsmanOracle, 5)
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	advanceToMainOf(t, g, seat)
	hand := handSize(me)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(pw), Strict: true}); err == nil {
		t.Fatal("−3 targeting a planeswalker was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: cardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == victim {
			t.Fatal("the creature survived")
		}
	}
	if handSize(me) != hand+1 {
		t.Errorf("hand = %d, want %d", handSize(me), hand+1)
	}
}

func TestGarrukCursedHuntsmanEmblemPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, me.ID, "Garruk, Cursed Huntsman", garrukCursedHuntsmanOracle, 6)
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Theirs")
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblemsOf(g, me.ID)))
	}
	if p, tough := effectivePTOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("my creature = %d/%d, want 5/5", p, tough)
	}
	if !hasKeywordOf(g, mine, "trample") {
		t.Error("my creature lacks trample")
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 || hasKeywordOf(g, theirs, "trample") {
		t.Error("the opponent's creature was affected")
	}
}
