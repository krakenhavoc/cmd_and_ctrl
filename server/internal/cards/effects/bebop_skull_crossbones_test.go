package effects

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const bebopSkullCrossbonesOracle = "59a405d7-2f28-41b8-a5bb-011f45a5fdf2"

// bebopTable seeds Bebop (2/1) for seat 0 with the given counters.
func bebopTable(t *testing.T, counters map[string]int) (g *game.Game, me *game.Player, bebop uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	bebop = pushCatalogPermanent(g, me.ID, "Bebop, Skull & Crossbones", "Legendary Creature — Boar Mutant", bebopSkullCrossbonesOracle, false)
	c := battlefieldCardPtr(t, g, bebop)
	c.Power, c.Toughness = 2, 1
	for kind, n := range counters {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(bebop, kind, n) })
	}
	return g, me, bebop
}

// bebopToPrompt passes priority until the Bebop trigger has resolved and
// asked, or the stack is empty.
func bebopToPrompt(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID); c != nil {
			return c
		}
		if stackFullyEmpty(g) {
			return nil
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
	return nil
}

func TestBebopSkullCrossbonesIsFullWithPartnerAndDeathtouch(t *testing.T) {
	spec, ok := Lookup(bebopSkullCrossbonesOracle)
	if !ok {
		t.Fatal("Bebop is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if !slices.Contains(spec.PrintedKeywords, "deathtouch") {
		t.Errorf("printed keywords %v, want deathtouch", spec.PrintedKeywords)
	}
	partner := 0
	for _, row := range spec.Triggered {
		if IsPartnerWithRow(row) {
			partner++
			if !strings.Contains(row.Key, "Rocksteady, Mutant Marauder") {
				t.Errorf("partner-with row %q does not name Rocksteady", row.Key)
			}
		}
	}
	if partner != 1 {
		t.Errorf("%d partner-with rows, want 1", partner)
	}
}

// X counts every counter on Bebop, of every kind: two +1/+1 counters and
// an oil counter are three cards and three life.
func TestBebopSkullCrossbonesDrawsAndLosesXOnAYes(t *testing.T) {
	g, me, bebop := bebopTable(t, map[string]int{game.CounterPlusOne: 2, "oil": 1})
	hand, life := me.Hand.Size(), me.Life
	attackWith(t, g, g.Seats[1].ID, bebop)
	c := bebopToPrompt(t, g, me)
	if c == nil {
		t.Fatal("Bebop did not ask")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want three cards", hand, me.Hand.Size())
	}
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want three lost", life, me.Life)
	}
}

func TestBebopSkullCrossbonesDeclineDoesNothing(t *testing.T) {
	g, me, bebop := bebopTable(t, map[string]int{game.CounterPlusOne: 2})
	hand, life := me.Hand.Size(), me.Life
	attackWith(t, g, g.Seats[1].ID, bebop)
	c := bebopToPrompt(t, g, me)
	if c == nil {
		t.Fatal("Bebop did not ask")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d, want both unchanged", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestBebopSkullCrossbonesWithNoCountersDrawsNothing(t *testing.T) {
	g, me, bebop := bebopTable(t, nil)
	hand, life := me.Hand.Size(), me.Life
	attackWith(t, g, g.Seats[1].ID, bebop)
	if c := bebopToPrompt(t, g, me); c != nil {
		t.Fatalf("Bebop asked with no counters: %+v", c)
	}
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d, want both unchanged", hand, me.Hand.Size(), life, me.Life)
	}
}

// X is read as the trigger resolves: a counter added in response counts.
func TestBebopSkullCrossbonesCountsCountersAsItResolves(t *testing.T) {
	g, me, bebop := bebopTable(t, map[string]int{game.CounterPlusOne: 1})
	hand := me.Hand.Size()
	attackWith(t, g, g.Seats[1].ID, bebop)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bebop, game.CounterPlusOne, 1) })
	c := bebopToPrompt(t, g, me)
	if c == nil {
		t.Fatal("Bebop did not ask")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want two cards", hand, me.Hand.Size())
	}
}

// A Bebop that left in response is read as it last existed (CR 608.2h).
func TestBebopSkullCrossbonesUsesLastKnownCountersWhenItLeft(t *testing.T) {
	g, me, bebop := bebopTable(t, map[string]int{game.CounterPlusOne: 2})
	life := me.Life
	attackWith(t, g, g.Seats[1].ID, bebop)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bebop) })
	hand := me.Hand.Size()
	c := bebopToPrompt(t, g, me)
	if c == nil {
		t.Fatal("Bebop did not ask after leaving")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("hand %d -> %d, life %d -> %d, want two cards and two life", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestBebopSkullCrossbonesIgnoresNoncombatDamage(t *testing.T) {
	g, me, bebop := bebopTable(t, map[string]int{game.CounterPlusOne: 2})
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bebop, g.Seats[1].ID, 2) })
	if c := bebopToPrompt(t, g, me); c != nil {
		t.Fatalf("Bebop asked after noncombat damage: %+v", c)
	}
}
