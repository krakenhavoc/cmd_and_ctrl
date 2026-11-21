package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// upsTapped parks a tapped permanent.
func UpsTapped(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Tapped: true,
	})
	return id
}

func UpsIsTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func upsSeatOf(g *game.Game, id uuid.UUID) int {
	for i, p := range g.Seats {
		if p != nil && p.ID == id {
			return i
		}
	}
	return -1
}

// upsAfterTheirUntap walks to the upkeep of `player`'s next turn, so
// their untap step has just happened.
func UpsAfterTheirUntap(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	seat := upsSeatOf(g, player)
	if g.Turn.ActiveSeat == seat {
		// Walk off this turn first: its untap step is behind it.
		for g.Turn.ActiveSeat == seat {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
	}
	AdvanceToUpkeepOf(t, g, seat)
}
