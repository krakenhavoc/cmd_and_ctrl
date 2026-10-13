package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_flicker_test.go — #3159: a host that leaves the
// battlefield and comes back inside ONE resolution is a new object
// (CR 400.7) and no state-based action runs in between. Its Auras go to
// the graveyard (CR 303.4c, 704.5m) and its Equipment is unattached
// (CR 301.5c, 704.5n).
//
// Cloudshift and Ephemerate mint a new instance ID on the way back
// (resetAsNewObjectLocked), so the dangling link was always swept and
// these two pass without the epoch pin; they are the regression tests
// for the card shape. The same-ID path, where the pin is what does the
// work, is game/attach_new_object_test.go.

// soleBattlefieldIDNamed finds the one battlefield card with the name.
func soleBattlefieldIDNamed(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	ids := battlefieldIDsNamed(g, name)
	if len(ids) != 1 {
		t.Fatalf("battlefield has %d cards named %q, want 1", len(ids), name)
	}
	return ids[0]
}

func graveHasNamed(g *game.Game, owner uuid.UUID, name string) bool {
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

func TestFlickeredHostLosesItsAura(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, bear)
	if got := auraRestrictions(t, g, bear); got&game.CantAttackOrBlock == 0 {
		t.Fatalf("setup: Pacifism did not take hold (restrictions %v)", got)
	}

	castCatalogSpell(t, g, "Cloudshift", "Instant", cloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Pacifism", me.ID); n != 0 {
		t.Errorf("Pacifism stayed on the battlefield after its host was flickered (%d copies)", n)
	}
	if !graveHasNamed(g, me.ID, "Pacifism") {
		t.Error("Pacifism should be in its owner's graveyard (CR 303.4c)")
	}
	back := soleBattlefieldIDNamed(t, g, "Bear")
	if got := auraRestrictions(t, g, back); got&game.CantAttackOrBlock != 0 {
		t.Errorf("the returned creature is still pacified (restrictions %v)", got)
	}
}

func TestFlickeredHostLosesItsEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
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

	castCatalogSpell(t, g, "Cloudshift", "Instant", cloudshiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if n := namedOnBattlefield(g, "Bonesplitter", me.ID); n != 1 {
		t.Fatalf("Equipment should stay on the battlefield (CR 301.5c), found %d", n)
	}
	if host := attachmentHostOf(t, g, sword); host.Kind != "" {
		t.Errorf("Bonesplitter is still attached to %+v after its host was flickered", host)
	}
	back := soleBattlefieldIDNamed(t, g, "Bear")
	if got := effectivePower(t, g, back); got != 2 {
		t.Errorf("the returned creature kept the equipment bonus: power %d, want 2", got)
	}
}
