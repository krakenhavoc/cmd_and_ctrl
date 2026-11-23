package testsex

import (
	"testing"

	"github.com/google/uuid"

	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felidar Umbra (#3654): lifelink, {1}{W} attaches it to a creature you
// control, and umbra armor.

func TestFelidarUmbraGivesLifelinkAndMovesToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	other := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Felidar Umbra", cardsex.FelidarUmbraOracleID, bear)
	effectstest.AssertEffectiveKeyword(t, g, bear, "lifelink")

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other)}); err != nil {
		t.Fatalf("activate the attach ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Felidar Umbra", cardsex.FelidarUmbraOracleID, bear)
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
