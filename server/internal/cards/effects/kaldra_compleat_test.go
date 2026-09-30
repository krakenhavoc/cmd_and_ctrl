package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emitDamageEventForTest emits the damage event the combat damage step
// would, from `source` to `target`.
func emitDamageEventForTest(g *game.Game, source, target uuid.UUID, amount int, combat bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventDealDamage, Source: source, Target: target,
			Amount: amount, Combat: combat,
		})
	})
}

// The granted clause: the Germ deals combat damage to a creature that
// survives it (indestructible), and that creature is exiled.
func TestKaldraCompleatExilesACreatureItsHostDealsCombatDamageTo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	_, germ := castLivingWeapon(t, g, "Kaldra Compleat", kaldraCompleatOracle)
	// The emitted event marks no damage, so the wall survives it the
	// way an indestructible or high-toughness blocker would.
	blocker := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 20)

	emitDamageEventForTest(g, germ, blocker, 5, true)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(blocker) {
		t.Fatal("the creature the equipped Germ dealt combat damage to is still on the battlefield")
	}
	if !g.Exile.Contains(blocker) {
		t.Error("that creature is exiled")
	}
}

// Noncombat damage, and combat damage to a player, do not trigger it.
func TestKaldraCompleatIgnoresNoncombatDamageAndPlayers(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	_, germ := castLivingWeapon(t, g, "Kaldra Compleat", kaldraCompleatOracle)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 20)

	emitDamageEventForTest(g, germ, bear, 1, false)
	emitDamageEventForTest(g, germ, opp.ID, 5, true)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("noncombat damage exiled the creature")
	}
}

// The grant is the equipped creature's: an unequipped creature's
// combat damage exiles nothing.
func TestKaldraCompleatOnlyGrantsToTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castLivingWeapon(t, g, "Kaldra Compleat", kaldraCompleatOracle)
	other := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 20)

	emitDamageEventForTest(g, other, bear, 2, true)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("an unequipped creature's combat damage exiled a creature")
	}
}
