package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestBlightXViewCarriesTheCeilingAndTheCreatures is the #2174 view
// half: a mandatory "blight X" opens the X prompt (demands_x), names
// the greatest-toughness ceiling the engine enforces, and lists the
// creatures that could take the counters.
func TestBlightXViewCarriesTheCeilingAndTheCreatures(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-blight-x"
	prev := game.CatalogAdditionalCost
	game.CatalogAdditionalCost = func(key string) *game.AdditionalCost {
		if key == oracle {
			return &game.AdditionalCost{BlightX: true, Label: "Blight X"}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogAdditionalCost = prev })
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Blight Thing", TypeLine: "Sorcery", OracleID: oracle,
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	ac := c.AdditionalCost
	if ac == nil || !ac.BlightX || !ac.DemandsX {
		t.Fatalf("additional_cost = %+v, want blight_x and demands_x", ac)
	}
	if ac.BlightXMax != 0 || ac.BlightOptions == nil || len(ac.BlightOptions.Cards) != 0 {
		t.Errorf("no creature: max %d options %+v, want 0 and present-and-empty", ac.BlightXMax, ac.BlightOptions)
	}

	pushViewCreature(g, me.ID, 1, false) // a 1/2
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	ac = c.AdditionalCost
	var want int
	g.WithWriteLock(func() { want = g.BlightXCeilingForEffect(me.ID) })
	if want != 2 || ac.BlightXMax != want || len(ac.BlightOptions.Cards) != 1 {
		t.Errorf("one 2-toughness creature: max %d (engine %d) options %v, want 2 and one creature",
			ac.BlightXMax, want, ac.BlightOptions.Cards)
	}
}
