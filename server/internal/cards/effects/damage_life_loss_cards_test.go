package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_life_loss_cards_test.go is #2105 through the cards: damage
// that costs no life (infect, CR 702.90b; a life total that can't
// change, CR 119.8) is not a life loss to any of the three kinds of
// reader — an opponent's loss (Exquisite Blood, b04OpponentLostLife),
// your own loss (Vilis, s22PlayerLostLife), and the turn tally behind
// every "lost life this turn" card (b18LifeLostThisTurn) — while
// ordinary damage still is. Then Lich's Mastery (#2065), which the fix
// unblocked.

const lichsMasteryOracle = "785c306c-471c-4699-8ab2-43c253d569cf"

// --- an opponent's loss: Exquisite Blood -------------------------------

func TestExquisiteBloodIgnoresDamageThatCostsNoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Exquisite Blood", "Enchantment", b04ExquisiteBloodOracle, false)
	agent := b12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", blightedAgentOracle, 1, 1)
	before := me.Life

	// Infect: poison, no life lost, no gain.
	pr7Hit(t, g, agent, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if poisonOn(opp) != 2 || opp.Life != game.StartingLife {
		t.Fatalf("opponent: %d poison, %d life; want 2 and %d", poisonOn(opp), opp.Life, game.StartingLife)
	}
	if me.Life != before {
		t.Errorf("Exquisite Blood gained %d for infect damage, want 0", me.Life-before)
	}

	// A locked life total: the damage is dealt, nothing is lost.
	pushPermanentForTest(g, opp.ID, "Platinum Emperion", platinumEmperionOracle, "Artifact Creature — Golem")
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	pr7Hit(t, g, bear, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("Exquisite Blood gained %d for damage to a locked life total, want 0", me.Life-before)
	}

	// Ordinary damage to another opponent still counts.
	pr7Hit(t, g, bear, g.Seats[2].ID, 3)
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("after ordinary damage: gained %d, want 3", me.Life-before)
	}
}

// --- your own loss: Vilis ----------------------------------------------

func TestVilisDrawsNothingForDamageThatCostsNoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = pushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood", "Legendary Creature — Demon", vilisOracle, false)
	for range 10 {
		pushLibraryTopForTest(me, "Filler")
	}
	agent := b12Push(g, opp.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", blightedAgentOracle, 1, 1)
	bear := pr7Creature(g, opp.ID, "Bear", 2, "G")

	hand := me.Hand.Size()
	pr7Hit(t, g, agent, me.ID, 3)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("Vilis drew %d for infect damage, want 0", got)
	}

	// Ordinary damage: draws that many.
	pr7Hit(t, g, bear, me.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("Vilis drew %d for 2 ordinary damage, want 2", got)
	}

	// Locked: no draw.
	pushPermanentForTest(g, me.ID, "Platinum Emperion", platinumEmperionOracle, "Artifact Creature — Golem")
	hand = me.Hand.Size()
	pr7Hit(t, g, bear, me.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("Vilis drew %d for damage to a locked life total, want 0", got)
	}
}

// --- the turn tally: "lost life this turn" -----------------------------

func TestLifeLostThisTurnCountsOnlyLifeActuallyLost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	agent := b12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", blightedAgentOracle, 1, 1)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	pushPermanentForTest(g, third.ID, "Platinum Emperion", platinumEmperionOracle, "Artifact Creature — Golem")

	pr7Hit(t, g, agent, opp.ID, 3)
	pr7Hit(t, g, bear, third.ID, 4)
	g.ReadSnapshot(func() {
		if got := b18LifeLostThisTurn(g, opp.ID); got != 0 {
			t.Errorf("infect: life lost this turn = %d, want 0", got)
		}
		if got := b18LifeLostThisTurn(g, third.ID); got != 0 {
			t.Errorf("locked: life lost this turn = %d, want 0", got)
		}
	})
	pr7Hit(t, g, bear, opp.ID, 2)
	g.ReadSnapshot(func() {
		if got := b18LifeLostThisTurn(g, opp.ID); got != 2 {
			t.Errorf("ordinary: life lost this turn = %d, want 2", got)
		}
	})
}

// --- Lich's Mastery -----------------------------------------------------

func TestLichsMasteryDrawsForLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", lichsMasteryOracle, false)
	for range 10 {
		pushLibraryTopForTest(me, "Filler")
	}
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d for 3 life gained, want 3", got)
	}
}

// For each 1 life lost, one object from one mixed pool: here one
// permanent, one card from hand and one from the graveyard (the ruling).
func TestLichsMasteryExilesOnePerLifeLostFromAnyMix(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mastery := pushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", lichsMasteryOracle, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	inHand := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	_ = handCardForTest(me, "Island", "Basic Land — Island", "")
	inYard := pushGraveyardCardForTest(me, "Old Card")
	_ = pushGraveyardCardForTest(me, "Older Card")
	src := pr7Creature(g, opp.ID, "Source", 3, "R")

	pr7Hit(t, g, src, me.ID, 3)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no exile prompt after losing 3 life")
	}
	if pick.ChooseMin != 3 || pick.ChooseMax != 3 {
		t.Fatalf("prompt bounds %d..%d, want exactly 3", pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, bear, inHand, inYard)
	for _, id := range []uuid.UUID{bear, inHand, inYard} {
		g.ReadSnapshot(func() {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				t.Errorf("%v is not in exile", id)
			}
		})
	}
	if findBattlefieldCardForTest(g, mastery) == nil {
		t.Error("Lich's Mastery was exiled although it was not picked")
	}
	if me.Life != game.StartingLife-3 {
		t.Errorf("life = %d, want %d: Lich's Mastery replaces no loss", me.Life, game.StartingLife-3)
	}
}

// Fewer objects than life lost: everything goes, Lich's Mastery
// included, and its leaving loses the game (the rulings).
func TestLichsMasteryExilesEverythingWhenShortThenLoses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mastery := pushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", lichsMasteryOracle, false)
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Graveyard.Cards = nil
	})
	src := pr7Creature(g, opp.ID, "Source", 5, "R")

	pr7Hit(t, g, src, me.ID, 5)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, mastery) != nil {
		t.Fatal("Lich's Mastery is still on the battlefield with nothing else to exile")
	}
	if !me.Eliminated {
		t.Fatal("Lich's Mastery left the battlefield and its controller is still in the game")
	}
}

// Infect damage loses no life, so it exiles nothing (#2105).
func TestLichsMasteryExilesNothingForInfectDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", lichsMasteryOracle, false)
	_ = pr7Creature(g, me.ID, "Bear", 2, "G")
	agent := b12Push(g, opp.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", blightedAgentOracle, 1, 1)

	pr7Hit(t, g, agent, me.ID, 2)
	passPriorityAroundTable(t, g)
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatalf("an exile prompt for infect damage: %+v", pick)
	}
	if poisonOn(me) != 2 {
		t.Errorf("poison = %d, want 2", poisonOn(me))
	}
}

// "You can't lose the game": 0 or less life keeps you in.
func TestLichsMasteryKeepsYouInAtZeroLife(t *testing.T) {
	g := newTwoSeatCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", lichsMasteryOracle, false)
	g.WithWriteLock(func() { me.Life = -3 })
	checkState(t, g)
	if me.Eliminated {
		t.Fatal("lost the game at -3 life under Lich's Mastery")
	}
}
