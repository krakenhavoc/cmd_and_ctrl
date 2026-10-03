package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr10_cards_test.go — ADR 0108 PR 10 (#1889): the two cards that
// deal damage as though its source had wither or infect, through the
// catalog.

const (
	pr10EverlastingTorment = "de41081e-713c-4005-90f7-65d9426208f1"
	pr10PhyrexianUnlife    = "6598d988-d60e-4441-88b6-8d0995c4675a"
)

func pr10Creature(g *game.Game, owner uuid.UUID, name string, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test",
		Power: 2, Toughness: toughness, Colors: []string{"R"}, Owner: owner, Controller: owner})
}

// Everlasting Torment, cast and resolved: damage to a creature is -1/-1
// counters, a prevention shield prevents none of it, nobody gains life, and
// damage to a player is still life loss.
func TestEverlastingTormentAllThreeLines(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Everlasting Torment", "Enchantment", pr10EverlastingTorment, nil)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByName(g, "Everlasting Torment") {
		t.Fatal("Everlasting Torment did not resolve onto the battlefield")
	}
	src := pr10Creature(g, me.ID, "Pinger", 1)
	victim := pr10Creature(g, opp.ID, "Victim", 5)
	theirs := opp.Life
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, victim, 3, false, "Mending Hands") {
			t.Fatal("no shield")
		}
		if err := g.DealDamageToCreatureForEffect(src, victim, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 4)
	})
	if got := minusOneOn(t, g, victim); got != 3 {
		t.Errorf("-1/-1 counters on the victim = %d, want 3: wither, and damage can't be prevented", got)
	}
	if got := damageMarkedOn(g, victim); got != 0 {
		t.Errorf("damage marked = %d, want 0", got)
	}
	if opp.Life != theirs-2 || poisonOn(opp) != 0 {
		t.Errorf("opponent life %d poison %d, want %d and 0: wither is about creatures, and nobody gains life",
			opp.Life, poisonOn(opp), theirs-2)
	}
}

// Phyrexian Unlife, cast and resolved, walking its rulings: damage that
// takes you from 3 to -2 is life loss and you stay in the game; the next
// damage is poison; ten poison still loses.
func TestPhyrexianUnlifeRulings(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Phyrexian Unlife", "Enchantment", pr10PhyrexianUnlife, nil)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByName(g, "Phyrexian Unlife") {
		t.Fatal("Phyrexian Unlife did not resolve onto the battlefield")
	}
	src := pr10Creature(g, opp.ID, "Hitter", 2)
	stSetLife(t, g, me, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 5); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if me.Eliminated || me.Life != -2 || poisonOn(me) != 0 {
		t.Fatalf("eliminated %v life %d poison %d; want in the game at -2 with no poison", me.Eliminated, me.Life, poisonOn(me))
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 4); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if me.Eliminated || me.Life != -2 || poisonOn(me) != 4 {
		t.Fatalf("eliminated %v life %d poison %d; want in the game at -2 with 4 poison", me.Eliminated, me.Life, poisonOn(me))
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 6); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Fatalf("still in the game with %d poison: Phyrexian Unlife stops only the life loss", poisonOn(me))
	}
}

// "If you're at 0 or less life and Phyrexian Unlife leaves the
// battlefield, you'll lose the game."
func TestPhyrexianUnlifeLeavingAtZeroLoses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	unlife := pushCatalogPermanent(g, me.ID, "Phyrexian Unlife", "Enchantment", pr10PhyrexianUnlife, false)
	stSetLife(t, g, me, -1)
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("lost at -1 life with Phyrexian Unlife out")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, unlife); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Fatal("still in the game at -1 life after Phyrexian Unlife left")
	}
}
