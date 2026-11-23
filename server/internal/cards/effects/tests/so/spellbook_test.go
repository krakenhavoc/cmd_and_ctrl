package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func pushSpellbook(g *game.Game, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		OracleID:   effectstest.SpellbookOracle,
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
	g := effectstest.NewTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	pushSpellbook(g, active.ID)
	effectstest.FillHandTo(t, g, active, 10)
	effectstest.RunCleanup(t, g)
	if len(g.DiscardPending) != 0 {
		t.Errorf("Spellbook controller was prompted to discard: %+v", g.DiscardPending)
	}
}

// TestSpellbookIsPlayerScoped: an opponent's Spellbook must not lift
// the active player's cap.
func TestSpellbookIsPlayerScoped(t *testing.T) {
	g := effectstest.NewTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushSpellbook(g, other.ID)
	effectstest.FillHandTo(t, g, active, 10)
	effectstest.RunCleanup(t, g)
	if len(g.DiscardPending) == 0 {
		t.Errorf("an opponent's Spellbook wrongly lifted the active player's hand-size cap")
	}
}
