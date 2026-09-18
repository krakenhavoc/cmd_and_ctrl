package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// commander_identity_test.go — the enumerator half of #844. CR 903.4f:
// "any color in your commander's color identity" adds nothing for a
// seat with no commander, or one whose commander's colour identity is
// colourless. Tapping Command Tower for no mana is legal and pointless,
// so it is not a move: a bot that took it would lose a land for
// nothing.

const oracleCommandTower = "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"

// stampCommander rewrites the seat's command-zone commander. Passing an
// identity of nil with a Scryfall printing behind it is Kozilek: a real
// colour identity that names no colours.
func stampCommander(t *testing.T, p *game.Player, name string, imported bool, identity []string) {
	t.Helper()
	for i := range p.Command.Cards {
		if !p.Command.Cards[i].IsCommander {
			continue
		}
		c := &p.Command.Cards[i]
		c.Name = name
		c.ColorIdentity, c.Colors, c.ManaCost = identity, nil, ""
		if imported {
			c.ScryfallID, c.OracleID = uuid.NewString(), uuid.NewString()
		}
		return
	}
	t.Fatalf("no commander in %s's command zone", p.Name)
}

func TestIdentityManaIsNotAMoveWithoutAnIdentity(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	tower := battlefieldCard(g, active, game.Card{Name: "Command Tower", TypeLine: "Land", OracleID: oracleCommandTower})
	advanceTo(t, g, game.StepPrecombatMain)

	// The table's commander is a placeholder with no colour data at
	// all. That is a data gap, not a colourless commander, so the
	// printed five stand and the tap is a move.
	if !hasManaMoveFrom(legal.EnumerateFor(g, active.ID), tower) {
		t.Error("a commander with no colour data should leave the printed card alone")
	}

	// A real colourless commander: the identity exists and names no
	// colours, so the ability adds nothing.
	stampCommander(t, active, "Kozilek, Butcher of Truth", true, nil)
	if hasManaMoveFrom(legal.EnumerateFor(g, active.ID), tower) {
		t.Error("offered a Command Tower tap that adds no mana (colourless commander)")
	}

	// An identity with colours in it is a move again.
	stampCommander(t, active, "Azorius Commander", true, []string{"W", "U"})
	if !hasManaMoveFrom(legal.EnumerateFor(g, active.ID), tower) {
		t.Error("a two-colour commander should leave the tap on offer")
	}

	// And with no commander at all, the quality is undefined: no move.
	active.Command.Cards = nil
	if hasManaMoveFrom(legal.EnumerateFor(g, active.ID), tower) {
		t.Error("offered a Command Tower tap that adds no mana (no commander)")
	}
}
