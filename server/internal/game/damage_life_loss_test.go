package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_life_loss_test.go is #2105: damage dealt to a player is life
// LOSS only when it costs life (CR 119.2, CR 120.3a). Damage from a
// source with infect gives poison instead (CR 120.3b, CR 702.90b), and
// damage to a player whose life total can't change moves nothing
// (CR 119.8, ADR 0085). Both are still DEALT: the event fires with the
// full Amount. What changes is the life-loss reading, which every
// "whenever a player loses life" trigger and the turn tally's LifeLost
// take from Event.DamageLifeLoss.

// lastDamageTo is the last EventDealDamage to `target` since `from`.
func lastDamageTo(t *testing.T, g *Game, from int, target uuid.UUID) Event {
	t.Helper()
	var out Event
	found := false
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventDealDamage && ev.Target == target {
			out, found = ev, true
		}
	}
	if !found {
		t.Fatalf("no EventDealDamage to %v: the damage must still be dealt", target)
	}
	return out
}

func TestOrdinaryDamageToAPlayerIsLifeLost(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 3, 3)
	})

	before := len(g.Events)
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(opp.ID, atk, 3, "") })
	ev := lastDamageTo(t, g, before, opp.ID)
	if ev.Amount != 3 || ev.DamageLifeLoss() != 3 || ev.DamageNotLifeLoss != 0 {
		t.Errorf("combat event: Amount %d, life loss %d, not-lost %d; want 3, 3, 0",
			ev.Amount, ev.DamageLifeLoss(), ev.DamageNotLifeLoss)
	}

	before = len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(atk, opp.ID, 2); err != nil {
			t.Fatalf("noncombat damage: %v", err)
		}
	})
	if ev := lastDamageTo(t, g, before, opp.ID); ev.DamageLifeLoss() != 2 {
		t.Errorf("noncombat event life loss = %d, want 2", ev.DamageLifeLoss())
	}
	if got := g.TurnTallyFor(opp.ID).LifeLost; got != 5 {
		t.Errorf("tally LifeLost = %d, want 5", got)
	}
	if opp.Life != StartingLife-5 {
		t.Errorf("life = %d, want %d", opp.Life, StartingLife-5)
	}
}

// Toxic's poison is IN ADDITION to the life loss (CR 120.3g), so all
// the damage is still life lost.
func TestToxicDamageIsStillAllLifeLost(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, "toxic 1")
	})
	before := len(g.Events)
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(opp.ID, atk, 2, "") })
	if ev := lastDamageTo(t, g, before, opp.ID); ev.DamageLifeLoss() != 2 {
		t.Errorf("life loss = %d, want 2", ev.DamageLifeLoss())
	}
	if got := g.TurnTallyFor(opp.ID).LifeLost; got != 2 {
		t.Errorf("tally LifeLost = %d, want 2", got)
	}
}

// The probe from the issue: a 2-power infect source dealing 2 to a
// player at full life leaves them there with 2 poison, and no reader
// may count it as life lost.
func TestInfectDamageIsNotLifeLost(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() {
		pushPrintedKeywordCreature(g, atk, me.ID, me.ID, 2, 2, KeywordInfect)
	})

	before := len(g.Events)
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(opp.ID, atk, 2, "") })
	ev := lastDamageTo(t, g, before, opp.ID)
	if ev.Amount != 2 || ev.DamageLifeLoss() != 0 {
		t.Errorf("combat infect event: Amount %d, life loss %d; want 2 dealt and 0 lost", ev.Amount, ev.DamageLifeLoss())
	}

	before = len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(atk, opp.ID, 3); err != nil {
			t.Fatalf("noncombat infect damage: %v", err)
		}
	})
	if ev := lastDamageTo(t, g, before, opp.ID); ev.Amount != 3 || ev.DamageLifeLoss() != 0 {
		t.Errorf("noncombat infect event: Amount %d, life loss %d; want 3 dealt and 0 lost", ev.Amount, ev.DamageLifeLoss())
	}

	if got := g.TurnTallyFor(opp.ID).LifeLost; got != 0 {
		t.Errorf("tally LifeLost = %d, want 0: infect damage loses no life (CR 702.90b)", got)
	}
	if opp.Life != StartingLife || poisonOf(g, opp.ID) != 5 {
		t.Errorf("life %d poison %d, want %d and 5", opp.Life, poisonOf(g, opp.ID), StartingLife)
	}
}

// Phyrexian Unlife's "as though its source had infect" is the same
// result through the other door (ADR 0108 §10).
func TestDamageAsThoughInfectIsNotLifeLost(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	pushAsThoughEnchantment(g, me)
	setLife(g, me, 0)

	before := len(g.Events)
	dealToPlayer(t, g, src, me.ID, 3)
	if ev := lastDamageTo(t, g, before, me.ID); ev.DamageLifeLoss() != 0 {
		t.Errorf("life loss = %d, want 0", ev.DamageLifeLoss())
	}
	if got := g.TurnTallyFor(me.ID).LifeLost; got != 0 {
		t.Errorf("tally LifeLost = %d, want 0", got)
	}
}

// Platinum Emperion: the damage is dealt, the total does not move, and
// so no life was lost (CR 119.8, ADR 0085 Decision 5).
func TestDamageToALockedLifeTotalIsNotLifeLost(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lockUntilNextTurn(g, me)
	attacker := pushColouredCreature(g, opp, "Big Attacker", []string{"G"})

	before := len(g.Events)
	g.WithWriteLock(func() { g.markCombatDamageToPlayerLocked(me.ID, attacker, 4, "") })
	if ev := lastDamageTo(t, g, before, me.ID); ev.Amount != 4 || ev.DamageLifeLoss() != 0 {
		t.Errorf("combat event: Amount %d, life loss %d; want 4 dealt and 0 lost", ev.Amount, ev.DamageLifeLoss())
	}
	before = len(g.Events)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(attacker, me.ID, 3); err != nil {
			t.Fatalf("noncombat damage: %v", err)
		}
	})
	if ev := lastDamageTo(t, g, before, me.ID); ev.DamageLifeLoss() != 0 {
		t.Errorf("noncombat life loss = %d, want 0", ev.DamageLifeLoss())
	}
	if got := g.TurnTallyFor(me.ID).LifeLost; got != 0 {
		t.Errorf("tally LifeLost = %d, want 0", got)
	}
}

// An event written before the field existed (a restore point's queued
// trigger) reads as the life loss it was taken for then.
func TestDamageLifeLossDefaultsToTheDamage(t *testing.T) {
	ev := Event{Kind: EventDealDamage, Amount: 4}
	if got := ev.DamageLifeLoss(); got != 4 {
		t.Errorf("DamageLifeLoss = %d, want 4", got)
	}
	if got := (Event{Kind: EventChangeLife, Amount: -4}).DamageLifeLoss(); got != 0 {
		t.Errorf("a life change's DamageLifeLoss = %d, want 0", got)
	}
}
