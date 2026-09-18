package game

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// Regression tests for two turn-wrap bugs the S31 bot fuzzer found on
// its first run (S31 sub-PR 3). Both live at the cleanup → next-untap
// seam, which the ordinary priority-pass path takes every turn.

func newWrapGame(t *testing.T, seats int) *Game {
	t.Helper()
	g := NewGame()
	for i := 0; i < seats; i++ {
		deck := []Card{NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			c := NewCard(fmt.Sprintf("Mountain %d", j), uuid.Nil)
			c.TypeLine = "Basic Land — Mountain"
			deck = append(deck, c)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

// passUntilNewTurn passes priority until the active seat changes,
// failing if the cursor ever lands on a step with no priority holder
// (which is exactly the stall the eliminated-seat bug produced).
func passUntilNewTurn(t *testing.T, g *Game) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	for i := 0; i < 64; i++ {
		if g.Turn.ActiveSeat != seat {
			return
		}
		if g.Turn.Step == StepCleanup && len(g.DiscardPending) > 0 {
			// The S13.4 cleanup pause: discard down and let the hook
			// re-fire. Not the stall under test.
			for pid, n := range g.DiscardPending {
				p := g.PlayerByID(pid)
				ids := make([]uuid.UUID, 0, n)
				for _, c := range p.Hand.Cards[:n] {
					ids = append(ids, c.InstanceID)
				}
				if err := g.DiscardSelection(pid, ids); err != nil {
					t.Fatal(err)
				}
			}
			continue
		}
		if g.Turn.PriorityHolder == NoPriority {
			t.Fatalf("cursor parked with no priority at %s (turn %d seat %d)", g.Turn.Step, g.Turn.Number, g.Turn.ActiveSeat)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority at %s: %v", g.Turn.Step, err)
		}
	}
	t.Fatalf("turn never wrapped from seat %d", seat)
}

func TestPerTurnCachesClearOnPriorityWrap(t *testing.T) {
	g := newWrapGame(t, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != StepPrecombatMain {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	land := active.Hand.Cards[0]
	if err := g.CastSpell(active.ID, land.InstanceID, CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	if got := g.LandsPlayedThisTurnFor(active.ID); got != 1 {
		t.Fatalf("LandsPlayedThisTurn after a land drop: %d", got)
	}
	g.SpellsCastThisTurn = map[uuid.UUID]CastTally{active.ID: {Total: 1, Noncreature: 1}}
	g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{uuid.New(): true}

	passUntilNewTurn(t, g)

	if got := g.LandsPlayedThisTurnFor(active.ID); got != 0 {
		t.Errorf("LandsPlayedThisTurn survived the wrap: %d", got)
	}
	if len(g.SpellsCastThisTurn) != 0 {
		t.Errorf("SpellsCastThisTurn survived the wrap: %v", g.SpellsCastThisTurn)
	}
	if len(g.LoyaltyActivatedThisTurn) != 0 {
		t.Errorf("LoyaltyActivatedThisTurn survived the wrap: %v", g.LoyaltyActivatedThisTurn)
	}
}

func TestTurnRotationSkipsEliminatedSeats(t *testing.T) {
	g := newWrapGame(t, 4)
	// Seats 1 and 2 leave the game. Seat 0 is active on turn 1; the
	// rotation must go 0 → 3 → 0, never parking on 1 or 2.
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if g.CurrentState() != StateActive {
		t.Fatalf("two of four conceding should not end the game: %s", g.CurrentState())
	}
	want := []int{3, 0, 3, 0}
	for _, seat := range want {
		passUntilNewTurn(t, g)
		if g.Turn.ActiveSeat != seat {
			t.Fatalf("active seat %d, want %d (turn %d)", g.Turn.ActiveSeat, seat, g.Turn.Number)
		}
		if g.Seats[g.Turn.ActiveSeat].Eliminated {
			t.Fatalf("turn dealt to eliminated seat %d", g.Turn.ActiveSeat)
		}
		ph := g.Turn.PriorityHolder
		if ph < 0 || g.Seats[ph].Eliminated {
			t.Fatalf("priority on eliminated / no seat: %d", ph)
		}
	}
	// PassTurn takes the same seam.
	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.ActiveSeat != 3 {
		t.Errorf("PassTurn from seat 0 should land on seat 3, got %d", g.Turn.ActiveSeat)
	}
	// The turn number still advances when the wrap passes seat 0.
	if g.Turn.Number < 3 {
		t.Errorf("turn number should have advanced past the skipped seats: %d", g.Turn.Number)
	}
}

// TestDamageAssignmentZeroEntriesDoNotConstrainPriors: with three
// blockers and two power, [2, 0, 0] is the only legal split (CR
// 510.1c) and must be accepted; [1, 1, 0] must still be rejected.
func TestDamageAssignmentZeroEntriesDoNotConstrainPriors(t *testing.T) {
	g := newWrapGame(t, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[1-g.Turn.ActiveSeat]
	mk := func(owner *Player, name string, p, tgh int) uuid.UUID {
		c := NewCard(name, owner.ID)
		c.TypeLine = "Creature — Bear"
		c.Power, c.Toughness = p, tgh
		c.Controller = owner.ID
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	atk := mk(active, "Attacker", 2, 2)
	b1, b2, b3 := mk(def, "B1", 2, 2), mk(def, "B2", 2, 2), mk(def, "B3", 3, 3)
	for g.Turn.Step != StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.DeclareAttacker(atk, def.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	for _, b := range []uuid.UUID{b1, b2, b3} {
		if err := g.DeclareBlocker(b, atk); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceDamageAssignment {
		t.Fatalf("expected one damage-assignment prompt, got %d", len(g.PendingChoices))
	}
	id := g.PendingChoices[0].ID
	bad := []DamageAssignmentEntry{{BlockerID: b1, Amount: 1}, {BlockerID: b2, Amount: 1}, {BlockerID: b3, Amount: 0}}
	if err := g.ResolveDamageAssignment(id, active.ID, bad, 0); err == nil {
		t.Errorf("[1,1,0] should violate prefix-lethal")
	}
	good := []DamageAssignmentEntry{{BlockerID: b1, Amount: 2}, {BlockerID: b2, Amount: 0}, {BlockerID: b3, Amount: 0}}
	if err := g.ResolveDamageAssignment(id, active.ID, good, 0); err != nil {
		t.Fatalf("[2,0,0] must be accepted: %v", err)
	}
}

// TestEliminationDropsThePlayersPendingChoices: a prompt owed by a
// player who leaves the game must not outlive them, or the table is
// blocked behind a choice nobody can answer.
func TestEliminationDropsThePlayersPendingChoices(t *testing.T) {
	g := newWrapGame(t, 3)
	leaver := g.Seats[1]
	stayer := g.Seats[2]
	g.QueueChoiceForEffect(PendingChoice{Kind: PendingChoiceDiscardFromHand, Chooser: leaver.ID, FromPlayer: leaver.ID, Count: 1, Reason: "test"})
	g.QueueChoiceForEffect(PendingChoice{Kind: PendingChoiceDiscardFromHand, Chooser: stayer.ID, FromPlayer: stayer.ID, Count: 1, Reason: "test"})
	g.DiscardPending = map[uuid.UUID]int{leaver.ID: 2}
	if len(g.PendingChoices) != 2 {
		t.Fatalf("setup: %d choices", len(g.PendingChoices))
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != stayer.ID {
		t.Errorf("leaver's choice should be dropped and the stayer's kept: %+v", g.PendingChoices)
	}
	if _, owed := g.DiscardPending[leaver.ID]; owed {
		t.Errorf("leaver's cleanup discard should be dropped")
	}
}

// TestEliminatedOwnersCommanderNeedsNoPrompt: the CR 903.9 "send it
// to the command zone?" prompt is addressed to the commander's owner;
// when that owner has left the game the engine answers "no" inline
// rather than queueing a prompt nobody can answer.
//
// The seat is flipped by hand rather than through Concede: since
// CR 800.4a (#769) a real concede takes the commander out of the game
// along with everything else its owner owns, so the ordinary route to
// this board no longer exists. The gate is kept and pinned anyway —
// it is the last line of defence for any other way a permanent can
// end up owned by a seat that is no longer playing (a restored
// snapshot, a future rule that leaves one behind), and an unanswerable
// prompt wedges the whole table.
func TestEliminatedOwnersCommanderNeedsNoPrompt(t *testing.T) {
	g := newWrapGame(t, 3)
	leaver := g.Seats[1]
	cmdr := NewCommander("Leaver's Commander", leaver.ID)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.Power, cmdr.Toughness = 2, 2
	// Controlled by a seat that is still playing: a permanent left
	// under a DEPARTED player's control is exiled (CR 800.4c,
	// exileGhostControlledLocked), and this test is about the owner
	// being gone, not the controller.
	cmdr.Controller = g.Seats[0].ID
	g.Battlefield.PushTop(cmdr)
	g.WithWriteLock(func() { leaver.Eliminated = true })
	// Any player sacrificing it is fine for the test — the caller gate
	// is what SacrificePermanentForEffect skips.
	if err := g.SacrificePermanentForEffect(cmdr.InstanceID); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("prompt queued for an eliminated owner: %+v", g.PendingChoices[0])
	}
	if !leaver.Graveyard.Contains(cmdr.InstanceID) {
		t.Errorf("commander should have gone to the graveyard (the inline 'no'); command=%d graveyard=%d", leaver.Command.Size(), leaver.Graveyard.Size())
	}
}
