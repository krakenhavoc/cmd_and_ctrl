package game

import (
	"testing"

	"github.com/google/uuid"
)

// stale_commander_prompt_test.go pins #605: a CR 903.9 "send your
// commander to the command zone instead?" prompt must never outlive
// the card it asks about.
//
// The soak found a table wedged at turn 19 with 26 pending choices,
// twelve of them the same commander-zone question, every answer
// refused with "card instance not found in zone". Two halves to it:
//
//   - The SBA sweep re-doomed a commander whose exit was ALREADY
//     paused on the prompt. A paused exit moves nothing, so the
//     commander kept its zero toughness (or its deathtouch mark) and
//     the next sweep doomed it again — and because a destroy counts
//     as an SBA firing, runStateChecksLocked went round again, up to
//     its 32-iteration cap. One death, thirty-two prompts.
//   - The moment the first was answered the card left the
//     battlefield, and the other thirty-one became unanswerable: the
//     resume looks for the card on the battlefield and does not find
//     it. The seat was offered those dead answers and nothing else,
//     so one commander's death froze the whole table.
//
// The tests below pin one prompt per death, and pin the prune that
// catches the stale siblings if anything ever queues them again.

// seatBattlefieldCommander puts a 4/4 legendary commander on the
// battlefield and returns its instance ID. `mut` doctors the card
// before it lands — the two ways a permanent stays doomed across an
// SBA sweep are a toughness at or below zero and a deathtouch mark.
func seatBattlefieldCommander(t *testing.T, g *Game, owner *Player, mut func(c *Card)) uuid.UUID {
	t.Helper()
	id := uuid.New()
	c := Card{
		InstanceID:  id,
		Name:        "Test Commander",
		TypeLine:    "Legendary Creature — Angel",
		Power:       4,
		Toughness:   4,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: true,
	}
	if mut != nil {
		mut(&c)
	}
	g.Battlefield.PushTop(c)
	return id
}

// TestZeroToughnessCommanderQueuesOneCommandZonePrompt — a commander
// whose toughness hits zero (CR 704.5f) is asked about the command
// zone once, not once per state-check iteration.
func TestZeroToughnessCommanderQueuesOneCommandZonePrompt(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatBattlefieldCommander(t, g, owner, func(c *Card) {
		c.Counters = map[string]int{"-1/-1": 4}
	})

	g.mu.Lock()
	g.runStateChecksLocked()
	g.mu.Unlock()

	prompt := expectCommanderPrompt(t, g, owner)
	if !g.Battlefield.Contains(cmdID) {
		t.Fatal("the commander moved before its owner answered")
	}
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the answer", len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, g.Battlefield, g.Exile)
}

// TestDeathtouchedCommanderQueuesOneCommandZonePrompt — the other
// sticky doom. CR 702.2c's mark is only cleared at cleanup, so it
// survives the destroy attempt and, before #605, re-doomed the
// commander on every sweep just as a zero toughness did.
func TestDeathtouchedCommanderQueuesOneCommandZonePrompt(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatBattlefieldCommander(t, g, owner, func(c *Card) {
		c.MarkedLethalByDeathtouch = true
	})

	g.mu.Lock()
	g.runStateChecksLocked()
	g.mu.Unlock()

	prompt := expectCommanderPrompt(t, g, owner)
	// Declining is the other half of the question and must still
	// route the commander to its owner's graveyard.
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the answer", len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, owner.Graveyard, owner.Command, g.Battlefield, g.Exile)
}

// TestStaleCommandZonePromptIsPrunedWhenItsCardLeaves is the prune
// itself, driven from a board that has the duplicate prompts on it
// however they got there. Answering the first moves the commander,
// which must take its stale siblings with it.
func TestStaleCommandZonePromptIsPrunedWhenItsCardLeaves(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatBattlefieldCommander(t, g, owner, nil)

	// Two exits for one card, both paused — the shape the SBA sweep
	// used to produce on its own before the guard above existed.
	// Nothing has moved, so the second is a duplicate question about
	// a move that is already in flight.
	g.mu.Lock()
	for i := 0; i < 2; i++ {
		if err := g.routeBattlefieldCardToOwnerGraveyardLocked(cmdID); err != nil {
			t.Fatalf("routeBattlefieldCardToOwnerGraveyardLocked #%d: %v", i, err)
		}
		if !g.zoneChangePausedLocked(cmdID) {
			t.Fatal("a paused exit is not reported as paused")
		}
	}
	g.mu.Unlock()

	if len(g.PendingChoices) != 2 {
		t.Fatalf("expected the two duplicate prompts, got %d", len(g.PendingChoices))
	}
	first := g.PendingChoices[0]
	if err := g.ResolveOptionalReplacement(first.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the stale sibling survived its card leaving: %d prompts still queued",
			len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, g.Battlefield, g.Exile)
}

// TestStaleCommandZonePromptAnswerIsDroppedNotRefused is the answer
// path's own tolerance, which matters because a prompt the client
// already has in flight can race the prune. It must be dropped with a
// breadcrumb, not refused: a refusal leaves the seat holding a prompt
// it can never discharge, which is the wedge in #605.
func TestStaleCommandZonePromptAnswerIsDroppedNotRefused(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatBattlefieldCommander(t, g, owner, nil)

	g.mu.Lock()
	if err := g.routeBattlefieldCardToOwnerGraveyardLocked(cmdID); err != nil {
		t.Fatalf("routeBattlefieldCardToOwnerGraveyardLocked: %v", err)
	}
	// Move the commander by hand, behind the prune's back, so the
	// queued prompt is stale by the time it is answered.
	if _, err := MoveCard(g.Battlefield, owner.Command, cmdID); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	g.mu.Unlock()

	prompt := expectCommanderPrompt(t, g, owner)
	base := len(g.Events)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("answering a stale prompt was refused: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts still queued after the answer", len(g.PendingChoices))
	}
	// "No" on a live prompt sends the commander to the graveyard; on
	// a stale one the move must not happen a second time.
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, g.Battlefield, g.Exile)
	var sawBreadcrumb bool
	for _, ev := range g.Events[base:] {
		if ev.Kind == EventEffectError {
			sawBreadcrumb = true
			break
		}
	}
	if !sawBreadcrumb {
		t.Error("no EventEffectError breadcrumb for the dropped prompt")
	}
}
