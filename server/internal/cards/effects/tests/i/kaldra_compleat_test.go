package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// The granted clause: the Germ deals combat damage to a creature that
// survives it (indestructible), and that creature is exiled.
func TestKaldraCompleatExilesACreatureItsHostDealsCombatDamageTo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	_, germ := effectstest.CastLivingWeapon(t, g, "Kaldra Compleat", effectstest.KaldraCompleatOracle)
	// The emitted event marks no damage, so the wall survives it the
	// way an indestructible or high-toughness blocker would.
	blocker := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 20)

	effectstest.EmitDamageEventForTest(g, germ, blocker, 5, true)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(blocker) {
		t.Fatal("the creature the equipped Germ dealt combat damage to is still on the battlefield")
	}
	if !g.Exile.Contains(blocker) {
		t.Error("that creature is exiled")
	}
}

// Noncombat damage, and combat damage to a player, do not trigger it.
func TestKaldraCompleatIgnoresNoncombatDamageAndPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	_, germ := effectstest.CastLivingWeapon(t, g, "Kaldra Compleat", effectstest.KaldraCompleatOracle)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 20)

	effectstest.EmitDamageEventForTest(g, germ, bear, 1, false)
	effectstest.EmitDamageEventForTest(g, germ, opp.ID, 5, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("noncombat damage exiled the creature")
	}
}

// The grant is the equipped creature's: an unequipped creature's
// combat damage exiles nothing.
func TestKaldraCompleatOnlyGrantsToTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastLivingWeapon(t, g, "Kaldra Compleat", effectstest.KaldraCompleatOracle)
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 20)

	effectstest.EmitDamageEventForTest(g, other, bear, 2, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("an unequipped creature's combat damage exiled a creature")
	}
}
