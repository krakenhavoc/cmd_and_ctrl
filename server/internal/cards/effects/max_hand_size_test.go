package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Issue #338: Thought Vessel's "you have no maximum hand size" half
// shipped as a declared no-op, but hand-size enforcement landed in
// S13.4, so the claim went stale and a real game made a Thought
// Vessel controller discard at their end step. The oracle ID lives
// in aang_batch1_test.go as thoughtVesselOracle.

// pushThoughtVessel puts a Thought Vessel on the battlefield under
// `controller`'s control and returns its instance ID.
func pushThoughtVessel(g *game.Game, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		OracleID:   thoughtVesselOracle,
		Name:       "Thought Vessel",
		TypeLine:   "Artifact",
		Owner:      controller,
		Controller: controller,
	})
	return id
}

// fillHandTo draws until the player's hand holds at least n cards.
func fillHandTo(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < 60 && p.Hand.Size() < n; i++ {
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
	if p.Hand.Size() < n {
		t.Fatalf("hand only reached %d, want %d", p.Hand.Size(), n)
	}
}

// runCleanup walks the cursor to the end step and then takes the one
// AdvanceStep that enters cleanup, which is where
// populateDiscardPendingLocked decides who owes a discard.
func runCleanup(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into cleanup: %v", err)
	}
}

// leaveBattlefield sends a battlefield permanent to its owner's
// graveyard.
func leaveBattlefield(t *testing.T, g *game.Game, owner, cardID uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner},
		cardID,
	); err != nil {
		t.Fatalf("MoveCardByID off battlefield: %v", err)
	}
}

// TestCleanupDiscardControl is the control case: with nothing
// relevant on the battlefield, a 10-card hand DOES owe a cleanup
// discard. If this fails the rest of the file proves nothing.
func TestCleanupDiscardControl(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	fillHandTo(t, g, active, 10)
	runCleanup(t, g)
	if len(g.DiscardPending) == 0 {
		t.Fatalf("control: expected a discard prompt for a 10-card hand, got none")
	}
}

// TestThoughtVesselNoMaximumHandSize is issue #338 proper: a player
// with a Thought Vessel on the battlefield is not prompted to
// discard at cleanup with 8+ cards in hand.
func TestThoughtVesselNoMaximumHandSize(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	pushThoughtVessel(g, active.ID)
	fillHandTo(t, g, active, 10)
	runCleanup(t, g)
	if len(g.DiscardPending) != 0 {
		t.Errorf("Thought Vessel controller was prompted to discard: %+v",
			g.DiscardPending)
	}
}

// TestThoughtVesselIsPlayerScoped verifies the effect does not leak
// across the table: an opponent's Thought Vessel must not lift the
// active player's cap.
func TestThoughtVesselIsPlayerScoped(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushThoughtVessel(g, other.ID)
	fillHandTo(t, g, active, 10)
	runCleanup(t, g)
	if len(g.DiscardPending) == 0 {
		t.Errorf("an opponent's Thought Vessel wrongly lifted the " +
			"active player's hand-size cap")
	}
}

// TestTwoThoughtVesselsOneLeaves is the ownership case a
// "set the field on enter, restore it on leave" design gets wrong:
// with two on the battlefield, one leaving must not strand the
// player back at a maximum while the other is still there.
func TestTwoThoughtVesselsOneLeaves(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	first := pushThoughtVessel(g, active.ID)
	pushThoughtVessel(g, active.ID)
	fillHandTo(t, g, active, 10)

	leaveBattlefield(t, g, active.ID, first)

	runCleanup(t, g)
	if len(g.DiscardPending) != 0 {
		t.Errorf("one Thought Vessel leaving stranded the player at a "+
			"maximum while a second remained: %+v", g.DiscardPending)
	}
}

// TestThoughtVesselCapReturnsWhenLastLeaves is the other half: once
// the last one is gone the ordinary cap is back, with nothing left
// over needing to be restored.
func TestThoughtVesselCapReturnsWhenLastLeaves(t *testing.T) {
	g := newTestGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	only := pushThoughtVessel(g, active.ID)
	fillHandTo(t, g, active, 10)

	leaveBattlefield(t, g, active.ID, only)

	runCleanup(t, g)
	if len(g.DiscardPending) == 0 {
		t.Errorf("the hand-size cap did not come back after the last " +
			"Thought Vessel left the battlefield")
	}
}

// TestBirdsOfParadiseHasFlying is the second half of the #338
// stale-simplification sweep. Birds of Paradise shipped in S14 with
// a comment calling its flying "cosmetic (keyword enforcement is
// S18)" and no PrintedKeywords slot. S18 landed; the comment was
// never revisited, so the catalog entry blocked nothing.
//
// Read off-battlefield, which is the path a card that never went
// through deck import has to rely on.
func TestBirdsOfParadiseHasFlying(t *testing.T) {
	c := &game.Card{
		InstanceID: uuid.New(),
		OracleID:   "d3a0b660-358c-41bd-9cd2-41fbf3491b1a",
		TypeLine:   "Creature — Bird",
	}
	if !game.HasKeyword(c, "flying") {
		t.Errorf("Birds of Paradise: expected flying via the catalog")
	}
	if game.HasKeyword(c, "reach") {
		t.Errorf("Birds of Paradise: did not expect reach")
	}
}
