package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felidar Umbra (#3654): lifelink, {1}{W} attaches it to a creature you
// control, and umbra armor.

func TestFelidarUmbraGivesLifelinkAndMovesToAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	other := auraBear(g, me.ID)
	aura := auraCast(t, g, "Felidar Umbra", felidarUmbraOracleID, bear)
	assertEffectiveKeyword(t, g, bear, "lifelink")

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Targets: cardRefs(other)}); err != nil {
		t.Fatalf("activate the attach ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	var host uuid.UUID
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(aura)
		host = c.AttachedTo.ID
	})
	if host != other {
		t.Fatalf("Felidar Umbra is attached to %s, want the other creature %s", host, other)
	}
}

func TestFelidarUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Felidar Umbra", felidarUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
