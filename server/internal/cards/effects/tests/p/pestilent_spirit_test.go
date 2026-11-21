package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pestilent_spirit_test.go — the Sauron deck request (#2062): Pestilent
// Spirit's "instant and sorcery spells you control have deathtouch".

const pestilentSpiritOracle = "83930f98-5596-4713-998c-732d5ecedf72"

// boltABigCreature casts the active seat's Lightning Bolt at a fresh
// 5/5 the next seat controls, with a Pestilent Spirit under
// `spiritOwner` (uuid.Nil for none), and reports whether the 5/5 is
// still on the battlefield once the Bolt and the sweep after it are
// done.
func boltABigCreature(t *testing.T, spiritOwner func(g *game.Game) uuid.UUID) bool {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%len(g.Seats)]
	if spiritOwner != nil {
		effectstest.PushCatalogPermanent(g, spiritOwner(g), "Pestilent Spirit", "Creature — Spirit", pestilentSpiritOracle, false)
	}
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	return g.Battlefield.Contains(big)
}

func TestPestilentSpiritIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(pestilentSpiritOracle)
	if !ok {
		t.Fatal("Pestilent Spirit is not registered")
	}
	if spec.Name != "Pestilent Spirit" || spec.Completeness != effects.CompletenessFull {
		t.Errorf("registered as %q, %v; want Pestilent Spirit, Full", spec.Name, spec.Completeness)
	}
}

// The Bolt deals 3 to a 5/5 with deathtouch, so the creature is
// destroyed by CR 704.5h.
func TestPestilentSpiritGivesYourInstantDeathtouch(t *testing.T) {
	survived := boltABigCreature(t, func(g *game.Game) uuid.UUID { return g.Seats[g.Turn.ActiveSeat].ID })
	if survived {
		t.Error("a 5/5 survived 3 damage from an instant its controller's Pestilent Spirit gave deathtouch")
	}
}

// Without the Spirit, and with the Spirit on the other side of the
// table, the Bolt is just 3 damage and the 5/5 lives.
func TestPestilentSpiritGivesOtherPlayersSpellsNothing(t *testing.T) {
	if !boltABigCreature(t, nil) {
		t.Fatal("control: a 5/5 died to a plain Lightning Bolt")
	}
	survived := boltABigCreature(t, func(g *game.Game) uuid.UUID {
		return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	})
	if !survived {
		t.Error("an opponent's Pestilent Spirit gave the caster's Bolt deathtouch")
	}
}
