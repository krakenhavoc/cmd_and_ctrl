package protocol

import (
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Command Tower
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// identity_mana_view_test.go — the wire half of #844. CR 903.4f: with
// no commander, or a colourless one, Command Tower adds no mana, so its
// row ships adds_no_mana and the client greys it instead of letting a
// player tap the land for nothing. A card that still produces something
// never carries the flag.

func TestAddsNoManaTracksTheCommanderIdentity(t *testing.T) {
	const commandTowerOracle = "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"
	g := buildActiveGame(t)
	me := g.Seats[0]
	tower := game.NewCard("Command Tower", me.ID)
	tower.TypeLine = "Land"
	tower.OracleID = commandTowerOracle
	tower.Controller = me.ID
	g.Battlefield.PushTop(tower)
	forest := game.NewCard("Forest", me.ID)
	forest.TypeLine = "Basic Land — Forest"
	forest.Controller = me.ID
	g.Battlefield.PushTop(forest)

	towerRow := func(t *testing.T) ManaAbilityView {
		t.Helper()
		c := conditionCardView(t, g, tower.InstanceID)
		if len(c.ManaAbilities) != 1 {
			t.Fatalf("Command Tower mana abilities = %+v, want one", c.ManaAbilities)
		}
		return c.ManaAbilities[0]
	}

	// The test deck's commander is a placeholder with no colour data:
	// a data gap, so the printed card stands and the row is live.
	if towerRow(t).AddsNoMana {
		t.Error("a commander with no colour data must not switch the Tower off")
	}

	// A real colourless commander (Kozilek): the identity names no
	// colours, so the ability adds nothing.
	for i := range me.Command.Cards {
		if c := &me.Command.Cards[i]; c.IsCommander {
			c.Name = "Kozilek, Butcher of Truth"
			c.ScryfallID, c.OracleID = uuid.NewString(), uuid.NewString()
			c.ManaCost = "{10}"
		}
	}
	if !towerRow(t).AddsNoMana {
		t.Error("colourless commander: want adds_no_mana on the Command Tower row")
	}
	if b := conditionCardView(t, g, forest.InstanceID); len(b.ManaAbilities) == 0 || b.ManaAbilities[0].AddsNoMana {
		t.Errorf("a Forest's mana ability must never carry adds_no_mana: %+v", b.ManaAbilities)
	}

	// No commander at all — CR 903.4f's own case.
	me.Command.Cards = nil
	if !towerRow(t).AddsNoMana {
		t.Error("no commander: want adds_no_mana on the Command Tower row")
	}
}
