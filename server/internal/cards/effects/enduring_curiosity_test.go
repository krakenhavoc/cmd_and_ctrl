package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const enduringCuriosityOracle = "9d2460c3-8eeb-4f35-b6f6-748c478664c7"

// TestEnduringCuriosityDrawsOncePerCreatureDealingCombatDamage pins
// the "a creature", not "one or more creatures", shape: two separate
// creatures connecting is two separate draws, with no OncePerBatch
// collapsing them.
func TestEnduringCuriosityDrawsOncePerCreatureDealingCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle, false)
	bearA := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear A", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	bearB := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear B", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()

	dealCombatDamageToPlayer(g, bearA, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("one creature connecting: hand %d, want %d", got, before+1)
	}

	dealCombatDamageToPlayer(g, bearB, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Fatalf("a second creature connecting: hand %d, want %d", got, before+2)
	}
}

// TestEnduringCuriosityDiesReturnsAsANonCreatureEnchantment is the
// card's whole reason for being #321: dying as a creature returns it
// under its owner's control as a non-creature enchantment, and dying
// a SECOND time — now as an enchantment — does not trigger a second
// return, because CR 603.4's "if it was a creature" is checked
// against that death's own last-known information.
func TestEnduringCuriosityDiesReturnsAsANonCreatureEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cat := pushCatalogPermanent(g, me.ID, "Enduring Curiosity", "Enchantment Creature — Cat Glimmer", enduringCuriosityOracle, false)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cat) })
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cat) {
		t.Fatal("Enduring Curiosity should be back on the battlefield, not left in the graveyard")
	}
	var isCreature, isEnchantment bool
	var owner, controller uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cat {
				continue
			}
			isCreature = c.IsCreature()
			isEnchantment = c.IsEnchantment()
			owner, controller = c.Owner, c.Controller
		}
	})
	if isCreature {
		t.Error("it should not be a creature any more — '(it's not a creature)'")
	}
	if !isEnchantment {
		t.Error("it should still be an enchantment")
	}
	if owner != me.ID || controller != me.ID {
		t.Errorf("should return under its owner's control: owner=%s controller=%s, want both %s", owner, controller, me.ID)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cat) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cat) {
		t.Error("dying a second time, now as an enchantment, should not return it again")
	}
}
