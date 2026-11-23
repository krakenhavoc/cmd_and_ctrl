package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- an opponent's loss: Exquisite Blood -------------------------------

func TestExquisiteBloodIgnoresDamageThatCostsNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Exquisite Blood", "Enchantment", effectstest.B04ExquisiteBloodOracle, false)
	agent := effectstest.B12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", effectstest.BlightedAgentOracle, 1, 1)
	before := me.Life

	// Infect: poison, no life lost, no gain.
	effectstest.Pr7Hit(t, g, agent, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PoisonOn(opp) != 2 || opp.Life != game.StartingLife {
		t.Fatalf("opponent: %d poison, %d life; want 2 and %d", effectstest.PoisonOn(opp), opp.Life, game.StartingLife)
	}
	if me.Life != before {
		t.Errorf("Exquisite Blood gained %d for infect damage, want 0", me.Life-before)
	}

	// A locked life total: the damage is dealt, nothing is lost.
	effectstest.PushPermanentForTest(g, opp.ID, "Platinum Emperion", effectstest.PlatinumEmperionOracle, "Artifact Creature — Golem")
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.Pr7Hit(t, g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("Exquisite Blood gained %d for damage to a locked life total, want 0", me.Life-before)
	}

	// Ordinary damage to another opponent still counts.
	effectstest.Pr7Hit(t, g, bear, g.Seats[2].ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("after ordinary damage: gained %d, want 3", me.Life-before)
	}
}

// --- your own loss: Vilis ----------------------------------------------

func TestVilisDrawsNothingForDamageThatCostsNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = effectstest.PushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood", "Legendary Creature — Demon", effectstest.VilisOracle, false)
	for range 10 {
		effectstest.PushLibraryTopForTest(me, "Filler")
	}
	agent := effectstest.B12Push(g, opp.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", effectstest.BlightedAgentOracle, 1, 1)
	bear := effectstest.Pr7Creature(g, opp.ID, "Bear", 2, "G")

	hand := me.Hand.Size()
	effectstest.Pr7Hit(t, g, agent, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("Vilis drew %d for infect damage, want 0", got)
	}

	// Ordinary damage: draws that many.
	effectstest.Pr7Hit(t, g, bear, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("Vilis drew %d for 2 ordinary damage, want 2", got)
	}

	// Locked: no draw.
	effectstest.PushPermanentForTest(g, me.ID, "Platinum Emperion", effectstest.PlatinumEmperionOracle, "Artifact Creature — Golem")
	hand = me.Hand.Size()
	effectstest.Pr7Hit(t, g, bear, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("Vilis drew %d for damage to a locked life total, want 0", got)
	}
}

// --- the turn tally: "lost life this turn" -----------------------------

func TestLifeLostThisTurnCountsOnlyLifeActuallyLost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	agent := effectstest.B12Push(g, me.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", effectstest.BlightedAgentOracle, 1, 1)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.PushPermanentForTest(g, third.ID, "Platinum Emperion", effectstest.PlatinumEmperionOracle, "Artifact Creature — Golem")

	effectstest.Pr7Hit(t, g, agent, opp.ID, 3)
	effectstest.Pr7Hit(t, g, bear, third.ID, 4)
	g.ReadSnapshot(func() {
		if got := effects.B18LifeLostThisTurn(g, opp.ID); got != 0 {
			t.Errorf("infect: life lost this turn = %d, want 0", got)
		}
		if got := effects.B18LifeLostThisTurn(g, third.ID); got != 0 {
			t.Errorf("locked: life lost this turn = %d, want 0", got)
		}
	})
	effectstest.Pr7Hit(t, g, bear, opp.ID, 2)
	g.ReadSnapshot(func() {
		if got := effects.B18LifeLostThisTurn(g, opp.ID); got != 2 {
			t.Errorf("ordinary: life lost this turn = %d, want 2", got)
		}
	})
}

// --- Lich's Mastery -----------------------------------------------------

func TestLichsMasteryDrawsForLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", effectstest.LichsMasteryOracle, false)
	for range 10 {
		effectstest.PushLibraryTopForTest(me, "Filler")
	}
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d for 3 life gained, want 3", got)
	}
}

// For each 1 life lost, one object from one mixed pool: here one
// permanent, one card from hand and one from the graveyard (the ruling).
func TestLichsMasteryExilesOnePerLifeLostFromAnyMix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mastery := effectstest.PushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", effectstest.LichsMasteryOracle, false)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	inHand := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	_ = effectstest.HandCardForTest(me, "Island", "Basic Land — Island", "")
	inYard := effectstest.PushGraveyardCardForTest(me, "Old Card")
	_ = effectstest.PushGraveyardCardForTest(me, "Older Card")
	src := effectstest.Pr7Creature(g, opp.ID, "Source", 3, "R")

	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no exile prompt after losing 3 life")
	}
	if pick.ChooseMin != 3 || pick.ChooseMax != 3 {
		t.Fatalf("prompt bounds %d..%d, want exactly 3", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear, inHand, inYard)
	for _, id := range []uuid.UUID{bear, inHand, inYard} {
		g.ReadSnapshot(func() {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				t.Errorf("%v is not in exile", id)
			}
		})
	}
	if effectstest.FindBattlefieldCardForTest(g, mastery) == nil {
		t.Error("Lich's Mastery was exiled although it was not picked")
	}
	if me.Life != game.StartingLife-3 {
		t.Errorf("life = %d, want %d: Lich's Mastery replaces no loss", me.Life, game.StartingLife-3)
	}
}

// Fewer objects than life lost: everything goes, Lich's Mastery
// included, and its leaving loses the game (the rulings).
func TestLichsMasteryExilesEverythingWhenShortThenLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mastery := effectstest.PushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", effectstest.LichsMasteryOracle, false)
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Graveyard.Cards = nil
	})
	src := effectstest.Pr7Creature(g, opp.ID, "Source", 5, "R")

	effectstest.Pr7Hit(t, g, src, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, mastery) != nil {
		t.Fatal("Lich's Mastery is still on the battlefield with nothing else to exile")
	}
	if !me.Eliminated {
		t.Fatal("Lich's Mastery left the battlefield and its controller is still in the game")
	}
}

// Infect damage loses no life, so it exiles nothing (#2105).
func TestLichsMasteryExilesNothingForInfectDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", effectstest.LichsMasteryOracle, false)
	_ = effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	agent := effectstest.B12Push(g, opp.ID, "Blighted Agent", "Creature — Phyrexian Human Rogue", effectstest.BlightedAgentOracle, 1, 1)

	effectstest.Pr7Hit(t, g, agent, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		t.Fatalf("an exile prompt for infect damage: %+v", pick)
	}
	if effectstest.PoisonOn(me) != 2 {
		t.Errorf("poison = %d, want 2", effectstest.PoisonOn(me))
	}
}

// "You can't lose the game": 0 or less life keeps you in.
func TestLichsMasteryKeepsYouInAtZeroLife(t *testing.T) {
	g := effectstest.NewTwoSeatCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lich's Mastery", "Legendary Enchantment", effectstest.LichsMasteryOracle, false)
	g.WithWriteLock(func() { me.Life = -3 })
	effectstest.CheckState(t, g)
	if me.Eliminated {
		t.Fatal("lost the game at -3 life under Lich's Mastery")
	}
}
