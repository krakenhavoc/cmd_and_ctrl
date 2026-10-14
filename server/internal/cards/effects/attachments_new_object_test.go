package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_new_object_test.go — #3159. A host that leaves the
// battlefield and comes back inside ONE resolution is a new object
// (CR 400.7), and no state-based check runs in between. Its Auras go to
// the graveyard (CR 303.4c, 704.5m) and its Equipment becomes
// unattached and stays (CR 301.5c, 704.5n).
//
// Chaos Warp is the catalogued card that brings a permanent back with
// the SAME instance ID: an empty library makes the tucked creature the
// card the reveal finds, and the library-to-battlefield put keeps the
// ID. That is the case Card.AttachedEpoch's pin decides.
//
// Cloudshift and Ephemerate go through the exile return, which mints a
// new instance ID (resetAsNewObjectLocked), so the old link names an
// object that no longer exists anywhere and the state-based action
// always swept it. Those two are regression tests for the card shape
// the issue names, not for the pin.

func graveHasName(g *game.Game, owner uuid.UUID, name string) bool {
	for _, p := range g.Seats {
		if p.ID != owner {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.Name == name {
				return true
			}
		}
	}
	return false
}

func TestChaosWarpedHostReturningAsTheSameIDLosesItsAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, bear)
	if got := auraRestrictions(t, g, bear); got&game.CantAttackOrBlock == 0 {
		t.Fatalf("setup: Pacifism did not take hold (restrictions %v)", got)
	}

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("setup: the warped creature is the only card in the library and comes back with its ID")
	}
	if n := namedOnBattlefield(g, "Pacifism", me.ID); n != 0 {
		t.Errorf("Pacifism followed the new object back onto the battlefield (%d copies)", n)
	}
	if !graveHasName(g, me.ID, "Pacifism") {
		t.Error("Pacifism should be in its owner's graveyard (CR 303.4c)")
	}
	if got := auraRestrictions(t, g, bear); got&game.CantAttackOrBlock != 0 {
		t.Errorf("the returned creature is still pacified (restrictions %v)", got)
	}
}

func TestChaosWarpedHostReturningAsTheSameIDLosesItsEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	advanceToMain(t, g)
	bear := auraBear(g, me.ID)
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	equipTo(t, g, me.ID, sword, bear)
	if host := attachmentHostOf(t, g, sword); host.ID != bear {
		t.Fatalf("setup: Bonesplitter not on the bear: %+v", host)
	}

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("setup: the warped creature is the only card in the library and comes back with its ID")
	}
	if host := attachmentHostOf(t, g, sword); host.Kind != "" {
		t.Errorf("Bonesplitter is still attached to %+v after its host became a new object (CR 301.5c)", host)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the returned creature kept the equipment bonus: power %d, want 2", got)
	}
}

func TestCloudshiftedHostLosesItsAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, bear)

	castCatalogSpell(t, g, "Cloudshift", "Instant", cloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Pacifism", me.ID); n != 0 {
		t.Errorf("Pacifism stayed on the battlefield after its host was flickered (%d copies)", n)
	}
	if !graveHasName(g, me.ID, "Pacifism") {
		t.Error("Pacifism should be in its owner's graveyard (CR 303.4c)")
	}
	back := battlefieldIDsNamed(g, "Bear")
	if len(back) != 1 {
		t.Fatalf("battlefield has %d Bears, want 1", len(back))
	}
	if got := auraRestrictions(t, g, back[0]); got&game.CantAttackOrBlock != 0 {
		t.Errorf("the returned creature is still pacified (restrictions %v)", got)
	}
}

func TestCloudshiftedHostLosesItsEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := auraBear(g, me.ID)
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	equipTo(t, g, me.ID, sword, bear)

	castCatalogSpell(t, g, "Cloudshift", "Instant", cloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bonesplitter", me.ID); n != 1 {
		t.Fatalf("Equipment should stay on the battlefield (CR 301.5c), found %d", n)
	}
	if host := attachmentHostOf(t, g, sword); host.Kind != "" {
		t.Errorf("Bonesplitter is still attached to %+v after its host was flickered", host)
	}
}
