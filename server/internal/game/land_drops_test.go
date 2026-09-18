package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// land_drops_test.go — #500. The per-turn land-play allowance used to
// be counted and never enforced: only the legal-move enumerator read
// the tally, so any client that didn't ask it could play a whole hand
// of lands on turn one. The owner's call was to enforce it, and to
// make the limit raisable rather than a literal 1 so Exploration and
// Azusa are a catalog line rather than an engine change.

// advanceToMain passes priority until the cursor reaches the active
// seat's precombat main phase — the only step a land may be played in.
func advanceToMainByPriority(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; g.Turn.Step != StepPrecombatMain; i++ {
		if i > 32 {
			t.Fatalf("never reached precombat main (stuck at %s)", g.Turn.Step)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority at %s: %v", g.Turn.Step, err)
		}
	}
}

// playLandFromHand plays the active seat's first hand card as a land.
func dropLandFromHand(t *testing.T, g *Game, p *Player) error {
	t.Helper()
	if len(p.Hand.Cards) == 0 {
		t.Fatal("hand is empty")
	}
	return g.CastSpell(p.ID, p.Hand.Cards[0].InstanceID, CastSpellParams{})
}

// withCatalogAdditionalLandPlays swaps the package-level hook for the
// duration of the test, in the shape withCatalogUntapStepPermissions
// uses.
func withCatalogAdditionalLandPlays(t *testing.T, fn func(oracleID string) int) {
	t.Helper()
	prev := CatalogAdditionalLandPlays
	CatalogAdditionalLandPlays = fn
	t.Cleanup(func() { CatalogAdditionalLandPlays = prev })
}

// TestSecondLandDropRefused is the bug in #500: the engine accepted
// land after land because nothing but the enumerator was counting.
func TestSecondLandDropRefused(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]

	if got := g.LandDropsRemainingFor(active.ID); got != 1 {
		t.Fatalf("land drops remaining before any play: %d, want 1", got)
	}
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("first land drop: %v", err)
	}
	if got := g.LandDropsRemainingFor(active.ID); got != 0 {
		t.Fatalf("land drops remaining after one play: %d, want 0", got)
	}

	handBefore := len(active.Hand.Cards)
	fieldBefore := len(g.Battlefield.Cards)
	err := dropLandFromHand(t, g, active)
	if !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("second land drop: err = %v, want ErrLandDropUnavailable", err)
	}
	// A refused play is a no-op: the card is still in hand, nothing
	// reached the battlefield, and the tally didn't move.
	if got := len(active.Hand.Cards); got != handBefore {
		t.Errorf("hand size after a refused land drop: %d, want %d", got, handBefore)
	}
	if got := len(g.Battlefield.Cards); got != fieldBefore {
		t.Errorf("battlefield size after a refused land drop: %d, want %d", got, fieldBefore)
	}
	if got := g.LandsPlayedThisTurnFor(active.ID); got != 1 {
		t.Errorf("LandsPlayedThisTurn after a refused land drop: %d, want 1", got)
	}
}

// TestLandDropAllowanceResetsEachTurn — the allowance is per turn, so
// the seat that spent it gets it back when its next turn begins.
func TestLandDropAllowanceResetsEachTurn(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	first := g.Seats[g.Turn.ActiveSeat]
	if err := dropLandFromHand(t, g, first); err != nil {
		t.Fatalf("first land drop: %v", err)
	}
	if err := dropLandFromHand(t, g, first); !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("second land drop on the same turn: err = %v, want ErrLandDropUnavailable", err)
	}

	// Round the table back to the same seat.
	passUntilNewTurn(t, g)
	passUntilNewTurn(t, g)
	if g.Seats[g.Turn.ActiveSeat].ID != first.ID {
		t.Fatalf("expected to be back on seat %d, got %d", first.Seat, g.Turn.ActiveSeat)
	}
	advanceToMainByPriority(t, g)

	if got := g.LandsPlayedThisTurnFor(first.ID); got != 0 {
		t.Fatalf("LandsPlayedThisTurn survived the turn: %d", got)
	}
	if got := g.LandDropsRemainingFor(first.ID); got != 1 {
		t.Fatalf("land drops remaining on the new turn: %d, want 1", got)
	}
	if err := dropLandFromHand(t, g, first); err != nil {
		t.Fatalf("land drop on the seat's next turn: %v", err)
	}
}

// TestRaisedBaseAllowanceAllowsExtraDrop — Player.LandDropsPerTurn is
// the base, and raising it raises the ceiling.
func TestRaisedBaseAllowanceAllowsExtraDrop(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	active.LandDropsPerTurn = 2

	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("first land drop: %v", err)
	}
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("second land drop with a base allowance of 2: %v", err)
	}
	if err := dropLandFromHand(t, g, active); !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("third land drop: err = %v, want ErrLandDropUnavailable", err)
	}
}

// TestOneTurnGrantAllowsExtraDropAndExpires — "you may play an
// additional land this turn" raises the ceiling for this turn only.
func TestOneTurnGrantAllowsExtraDropAndExpires(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]

	g.GrantAdditionalLandPlayForEffect(active.ID, 1)
	if got := g.LandDropsRemainingFor(active.ID); got != 2 {
		t.Fatalf("land drops remaining after a +1 grant: %d, want 2", got)
	}
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("first land drop: %v", err)
	}
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("second land drop under a +1 grant: %v", err)
	}
	if err := dropLandFromHand(t, g, active); !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("third land drop: err = %v, want ErrLandDropUnavailable", err)
	}

	passUntilNewTurn(t, g)
	if len(g.ExtraLandDropsThisTurn) != 0 {
		t.Errorf("one-turn land-play grant survived the turn: %v", g.ExtraLandDropsThisTurn)
	}
	if got := g.LandDropsRemainingFor(active.ID); got != 1 {
		t.Errorf("land drops for the granted seat next turn: %d, want 1", got)
	}
}

// TestControlledPermanentRaisesAllowance is the Exploration / Azusa
// shape: a static grant derived from the battlefield, so two of them
// compose and one leaving doesn't strand the other's grant.
func TestControlledPermanentRaisesAllowance(t *testing.T) {
	withCatalogAdditionalLandPlays(t, func(oracleID string) int {
		switch oracleID {
		case "exploration":
			return 1
		case "azusa":
			return 2
		}
		return 0
	})

	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	exploration := pushTappedPermanent(g, active.ID, "Exploration", "exploration", "Enchantment", false)
	if got := g.LandDropsRemainingFor(active.ID); got != 2 {
		t.Fatalf("land drops with one Exploration: %d, want 2", got)
	}
	// Someone else's Exploration is not yours.
	if got := g.LandDropsRemainingFor(other.ID); got != 1 {
		t.Errorf("opponent's land drops: %d, want 1", got)
	}
	// Two static grants compose.
	pushTappedPermanent(g, active.ID, "Azusa, Lost but Seeking", "azusa", "Legendary Creature — Human Monk", false)
	if got := g.LandDropsRemainingFor(active.ID); got != 4 {
		t.Fatalf("land drops with Exploration + Azusa: %d, want 4", got)
	}

	for i := 0; i < 4; i++ {
		if err := dropLandFromHand(t, g, active); err != nil {
			t.Fatalf("land drop %d of 4: %v", i+1, err)
		}
	}
	if err := dropLandFromHand(t, g, active); !errors.Is(err, ErrLandDropUnavailable) {
		t.Fatalf("fifth land drop: err = %v, want ErrLandDropUnavailable", err)
	}

	// One grant leaving leaves the other standing — the derivation
	// has no stored value to restore, which is the whole point.
	if _, err := MoveCard(g.Battlefield, active.Graveyard, exploration); err != nil {
		t.Fatal(err)
	}
	if got := g.EffectiveLandDropsLocked(active); got != 3 {
		t.Errorf("allowance after Exploration left: %d, want 3", got)
	}
}

// TestLandDropOnAnotherPlayersTurnRefused — enforcing the allowance
// must not have opened the timing gate. A land is a special action
// with sorcery timing (CR 305.1), so a non-active seat is refused and
// spends nothing.
func TestLandDropOnAnotherPlayersTurnRefused(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	handBefore := len(other.Hand.Cards)
	err := dropLandFromHand(t, g, other)
	if !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Fatalf("land drop on another player's turn: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if got := len(other.Hand.Cards); got != handBefore {
		t.Errorf("hand size after a refused out-of-turn land drop: %d, want %d", got, handBefore)
	}
	if got := g.LandsPlayedThisTurnFor(other.ID); got != 0 {
		t.Errorf("LandsPlayedThisTurn for a seat that never played: %d", got)
	}
}

// TestLandDropAllowanceSurvivesSnapshotRoundTrip — the base field and
// the one-turn grant both have to persist, and a pre-#500 snapshot
// (no base field, so zero) must restore as the default rather than as
// "may never play a land".
func TestLandDropAllowanceSurvivesSnapshotRoundTrip(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	active.LandDropsPerTurn = 2
	g.GrantAdditionalLandPlayForEffect(active.ID, 1)

	snap := g.CaptureSnapshot()
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got := restored.LandDropsRemainingFor(active.ID); got != 3 {
		t.Errorf("land drops after a round trip: %d, want 3", got)
	}

	// A snapshot from before the field existed: zero value everywhere.
	for i := range snap.Seats {
		snap.Seats[i].LandDropsPerTurn = 0
	}
	snap.ExtraLandDropsThisTurn = nil
	old, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore (pre-#500 shape): %v", err)
	}
	if got := old.LandDropsRemainingFor(active.ID); got != 1 {
		t.Errorf("land drops restored from a pre-#500 snapshot: %d, want 1", got)
	}
}

// TestGrantAdditionalLandPlayIgnoresNonsense — a grant is never a way
// to take land plays away, and a nil player is not a player.
func TestGrantAdditionalLandPlayIgnoresNonsense(t *testing.T) {
	g := newWrapGame(t, 2)
	active := g.Seats[g.Turn.ActiveSeat]

	g.GrantAdditionalLandPlayForEffect(active.ID, -3)
	g.GrantAdditionalLandPlayForEffect(active.ID, 0)
	g.GrantAdditionalLandPlayForEffect(uuid.Nil, 5)
	if len(g.ExtraLandDropsThisTurn) != 0 {
		t.Errorf("nonsense grants were recorded: %v", g.ExtraLandDropsThisTurn)
	}
	if got := g.LandDropsRemainingFor(active.ID); got != 1 {
		t.Errorf("land drops after nonsense grants: %d, want 1", got)
	}
}
