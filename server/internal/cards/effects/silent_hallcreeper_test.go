package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const silentHallcreeperOracle = "6aea681e-88d2-48df-a717-3ce0bc95205b"

// TestSilentHallcreeperIsUnblockable pins the static restriction.
func TestSilentHallcreeperIsUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Silent Hallcreeper", "Enchantment Creature — Horror", silentHallcreeperOracle, false)

	var restricted bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				restricted = c.Effective().Restrictions.Has(game.CantBeBlocked)
			}
		}
	})
	if !restricted {
		t.Error("Silent Hallcreeper should have CantBeBlocked")
	}
}

// TestSilentHallcreeperModePicksDrawACard — resolving the mode
// prompt with "Draw a card" draws a card after combat damage.
func TestSilentHallcreeperModePicksDrawACard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Silent Hallcreeper", OracleID: silentHallcreeperOracle,
		TypeLine: "Enchantment Creature — Horror", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	handBefore := me.Hand.Size()
	attackWith(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)

	mode := modePickChoiceFor(g, me.ID)
	if mode == nil {
		t.Fatalf("no mode_pick prompt for Silent Hallcreeper's trigger")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick (draw a card): %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand after choosing \"draw a card\": %d, want %d", got, handBefore+1)
	}
}
