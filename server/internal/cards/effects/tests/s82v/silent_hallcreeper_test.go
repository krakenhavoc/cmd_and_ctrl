package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestSilentHallcreeperIsUnblockable pins the static restriction.
func TestSilentHallcreeperIsUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Silent Hallcreeper", "Enchantment Creature — Horror", effectstest.SilentHallcreeperOracle, false)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Silent Hallcreeper", OracleID: effectstest.SilentHallcreeperOracle,
		TypeLine: "Enchantment Creature — Horror", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	handBefore := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)

	mode := effectstest.ModePickChoiceFor(g, me.ID)
	if mode == nil {
		t.Fatalf("no mode_pick prompt for Silent Hallcreeper's trigger")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick (draw a card): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand after choosing \"draw a card\": %d, want %d", got, handBefore+1)
	}
}
