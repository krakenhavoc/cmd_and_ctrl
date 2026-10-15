package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crab Umbra (#3654): {2}{U} untaps the enchanted creature, and umbra
// armor.

func TestCrabUmbraUntapsTheEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	aura := auraCast(t, g, "Crab Umbra", crabUmbraOracleID, bear)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the untap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tappedNow(g, bear) {
		t.Fatal("{2}{U}: Untap enchanted creature left it tapped")
	}
}

func TestCrabUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := umbraCreature(g, me.ID, 2, 2)
	umbra := auraCast(t, g, "Crab Umbra", crabUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
