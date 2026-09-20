package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spellbookOracle = "83133f51-bfbd-4db5-9be0-f660e3b66435"

func pushSpellbook(g *game.Game, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		OracleID:   spellbookOracle,
		Name:       "Spellbook",
		TypeLine:   "Artifact — Book",
		Owner:      controller,
		Controller: controller,
	})
	return id
}

// TestSpellbookLiftsTheHandSizeCap is the whole card: a ten-card
// hand walks through cleanup with a Spellbook out. The control case
// (a ten-card hand with nothing relevant on the battlefield DOES owe
// a discard) is TestCleanupDiscardControl in max_hand_size_test.go.
func TestSpellbookLiftsTheHandSizeCap(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	pushSpellbook(g, active.ID)
	fillHandTo(t, g, active, 10)
	runCleanup(t, g)
	if len(g.DiscardPending) != 0 {
		t.Errorf("Spellbook controller was prompted to discard: %+v", g.DiscardPending)
	}
}

// TestSpellbookIsPlayerScoped: an opponent's Spellbook must not lift
// the active player's cap.
func TestSpellbookIsPlayerScoped(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushSpellbook(g, other.ID)
	fillHandTo(t, g, active, 10)
	runCleanup(t, g)
	if len(g.DiscardPending) == 0 {
		t.Errorf("an opponent's Spellbook wrongly lifted the active player's hand-size cap")
	}
}
