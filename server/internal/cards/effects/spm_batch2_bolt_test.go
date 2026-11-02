package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const electrosBoltOracle = "e9222f89-e495-4502-bb0c-f5e802686b01"

// Electro's Bolt: 4 damage kills a 4-toughness creature; from the hand it
// costs {2}{R}, out of the graveyard only the mayhem {1}{R}.
func TestSpm2ElectrosBoltHandAndMayhem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	victim := seedCopyableCreature(g, g.Seats[1].ID, "Their Ox", "Creature — Ox", 2, 4)
	survivor := seedCopyableCreature(g, g.Seats[1].ID, "Their Wall", "Creature — Wall", 0, 5)
	bolt := mayhemHandCard(g, me, "Electro's Bolt", "Sorcery", "{2}{R}", electrosBoltOracle)

	floatForTest(g, me, "RR")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: survivor}}}); err == nil {
		t.Fatal("Electro's Bolt cast for two mana from the hand")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	mayhemDiscard(t, g, me, bolt)
	floatForTest(g, me, "RR")
	if err := castSlashFromGraveyard(g, me, bolt, victim, "mayhem"); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("4 damage left the 4-toughness creature alive")
	}
	if !g.Battlefield.Contains(survivor) {
		t.Error("the untargeted creature died")
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("a mayhem spell goes back to the graveyard")
	}
}

// Never discarded: the graveyard offer is refused.
func TestSpm2ElectrosBoltNotDiscardedIsNotCastable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	victim := seedCopyableCreature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	bolt := mayhemHandCard(g, me, "Electro's Bolt", "Sorcery", "{2}{R}", electrosBoltOracle)
	g.WithWriteLock(func() {
		_, _ = me.Hand.Remove(bolt)
		me.Library.PushTop(game.Card{InstanceID: bolt, Name: "Electro's Bolt", TypeLine: "Sorcery",
			ManaCost: "{2}{R}", OracleID: electrosBoltOracle, Owner: me.ID, Controller: me.ID})
		_ = g.MillNForEffect(me.ID, 1)
	})
	floatForTest(g, me, "RRR")
	if err := castSlashFromGraveyard(g, me, bolt, victim, "mayhem"); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("milled: err = %v, want ErrCastZoneNotAllowed", err)
	}
}
