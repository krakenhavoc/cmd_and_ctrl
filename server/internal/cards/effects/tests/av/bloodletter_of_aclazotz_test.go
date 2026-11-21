package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBloodletterDoublesNoncombatDamageToAnOpponentOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBloodletter(g, me.ID)
	theirs, mine := opp.Life, me.Life

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != theirs-6 {
		t.Errorf("a Bolt at an opponent on my turn: %d → %d, want %d", theirs, opp.Life, theirs-6)
	}
	ev, ok := effectstest.LastDamageTo(g, opp.ID)
	if !ok {
		t.Fatal("no EventDealDamage to the opponent")
	}
	if ev.Amount != 3 || ev.DamageLifeLoss() != 6 {
		t.Errorf("damage event: %d dealt, %d life lost; want 3 dealt and 6 lost (ADR 0169 §4)",
			ev.Amount, ev.DamageLifeLoss())
	}

	// Damage to me is not an opponent's loss.
	src := effectstest.PushVanillaCreature(g, opp.ID, "Their Pinger", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	if me.Life != mine-3 {
		t.Errorf("damage to Bloodletter's controller: %d → %d, want %d", mine, me.Life, mine-3)
	}
}

func TestBloodletterDoublesCombatDamageButNotLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBloodletter(g, me.ID)
	linker := effectstest.PushKeywordCreature(g, me.ID, "Linker", "lifelink")
	theirs, mine := opp.Life, me.Life

	effectstest.AttackWith(t, g, opp.ID, linker)

	if opp.Life != theirs-4 {
		t.Errorf("2 combat damage on my turn: %d → %d, want %d", theirs, opp.Life, theirs-4)
	}
	if me.Life != mine+2 {
		t.Errorf("lifelink gains the damage, not the loss: %d → %d, want %d", mine, me.Life, mine+2)
	}
}

func TestBloodletterIsOnlyYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBloodletter(g, me.ID)
	src := effectstest.PushVanillaCreature(g, me.ID, "My Pinger", 1, 1)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	theirs := opp.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2)
	})
	if opp.Life != theirs-5 {
		t.Errorf("3 damage and a 2-life drain on another player's turn: %d → %d, want %d",
			theirs, opp.Life, theirs-5)
	}
}

func TestBloodletterDoublesDrainsAndPaymentsButNotGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBloodletter(g, me.ID)

	theirs := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	if opp.Life != theirs-4 {
		t.Errorf("a 2-life drain: %d → %d, want %d", theirs, opp.Life, theirs-4)
	}

	theirs = opp.Life
	g.WithWriteLock(func() {
		if err := g.PayLifeForEffect(uuid.Nil, opp.ID, 2); err != nil {
			t.Fatalf("PayLifeForEffect: %v", err)
		}
	})
	if opp.Life != theirs-4 {
		t.Errorf("paying 2 life (CR 119.4): %d → %d, want %d", theirs, opp.Life, theirs-4)
	}

	theirs = opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	if opp.Life != theirs+3 {
		t.Errorf("gaining 3: %d → %d, want %d", theirs, opp.Life, theirs+3)
	}
}

func TestTwoBloodlettersQuadruple(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBloodletter(g, me.ID)
	effectstest.PushBloodletter(g, me.ID)
	src := effectstest.PushVanillaCreature(g, me.ID, "My Pinger", 1, 1)
	theirs := opp.Life

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 2) })

	if len(g.PendingChoices) != 0 {
		t.Fatalf("two copies of one replacement queued %d prompts, want 0", len(g.PendingChoices))
	}
	if opp.Life != theirs-8 {
		t.Errorf("2 damage under two Bloodletters: %d → %d, want %d", theirs, opp.Life, theirs-8)
	}
}

func TestBloodletterUnderALockedLifeTotalAndALifeFloor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, locked, floored := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushBloodletter(g, me.ID)
	src := effectstest.PushVanillaCreature(g, me.ID, "My Pinger", 1, 1)
	effectstest.PushPermanentForTest(g, locked.ID, "Platinum Emperion", effectstest.PlatinumEmperionOracle, "Artifact Creature — Golem")

	lockedLife := locked.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, locked.ID, 3) })
	if locked.Life != lockedLife {
		t.Errorf("a locked life total lost %d life", lockedLife-locked.Life)
	}

	floored.Life = 5
	g.WithWriteLock(func() {
		g.GrantLifeFloorForEffect(floored.ID, "Angel's Grace", uuid.Nil, game.Duration{})
		_ = g.DealDamageToPlayerForEffect(src, floored.ID, 3)
	})
	if floored.Life != 1 {
		t.Errorf("3 damage doubled to 6 from 5 under the floor: life %d, want 1", floored.Life)
	}
}

// TestBloodletterLifeLossTriggersSeeTheDoubledLoss — "whenever an
// opponent loses life" (Exquisite Blood) reads Event.DamageLifeLoss, so
// it sees the doubled loss of a damage event.
func TestBloodletterLifeLossTriggersSeeTheDoubledLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBloodletter(g, me.ID)
	effectstest.PushCatalogPermanent(g, me.ID, "Exquisite Blood", "Enchantment", effectstest.B04ExquisiteBloodOracle, false)
	mine := me.Life

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != mine+6 {
		t.Errorf("Exquisite Blood after a doubled Bolt: %d → %d, want %d", mine, me.Life, mine+6)
	}
}
