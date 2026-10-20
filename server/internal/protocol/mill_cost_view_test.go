package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keepLibraryCardsForTest trims `seat`'s library to at most n cards.
func keepLibraryCardsForTest(t *testing.T, g *game.Game, seat uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		p := g.PlayerByIDForEffect(seat)
		if p == nil {
			t.Fatal("no such seat")
		}
		if len(p.Library.Cards) > n {
			p.Library.Cards = p.Library.Cards[:n]
		}
	})
	g.BumpLayerVersionForTest()
}

// #3845: an activated row stamps its mill component, and a row whose
// controller's library is too short for it carries the engine's refusal
// as cant_activate, so the ready ring, the click and the engine agree.
func TestViewStampsMillCostAndShortfall(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var src uuid.UUID
	g.WithWriteLock(func() {
		src = uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: src, Name: "Mill Rows", TypeLine: "Artifact",
			OracleID: "00000000-0000-0000-0000-000000003845",
			Owner:    me.ID, Controller: me.ID,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "mill two", Cost: game.AbilityCost{Mill: 2}},
			},
		})
	})
	keepLibraryCardsForTest(t, g, me.ID, 1)

	rows := frameCard(t, g, me.ID.String(), src).ActivatedAbilities
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if r := rows[0]; r.MillCost != 2 || r.CantActivate != "Not enough cards in your library to mill (have 1, need 2)" {
		t.Errorf("short library: row = mill %d cant %q, want 2 and the shortfall", r.MillCost, r.CantActivate)
	}

	g.WithWriteLock(func() {
		g.PlayerByIDForEffect(me.ID).Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land"})
	})
	g.BumpLayerVersionForTest()
	if r := frameCard(t, g, me.ID.String(), src).ActivatedAbilities[0]; r.MillCost != 2 || r.CantActivate != "" {
		t.Errorf("two cards in the library: row = mill %d cant %q, want 2 and nothing", r.MillCost, r.CantActivate)
	}
}
