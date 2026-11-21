package testscr

import (
	"testing"

	"github.com/google/uuid"

	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crab Umbra (#3654): {2}{U} untaps the enchanted creature, and umbra
// armor.

func TestCrabUmbraUntapsTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	aura := effectstest.AuraCast(t, g, "Crab Umbra", cardsco.CrabUmbraOracleID, bear)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the untap ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, bear) {
		t.Fatal("{2}{U}: Untap enchanted creature left it tapped")
	}
}

func TestCrabUmbraReplacesADestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Crab Umbra", cardsco.CrabUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
