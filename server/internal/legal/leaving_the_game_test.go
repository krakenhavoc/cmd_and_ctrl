package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// leaving_the_game_test.go — #902. CR 800.4g hands a departed player's
// choice to somebody else, and the claim the engine makes is that the
// bot enumerator needs NOTHING for that: choiceMoves keys on
// PendingChoice.Chooser, so rewriting one field re-addresses the prompt
// for `internal/legal` exactly as it does for the client's view.
//
// This file is that claim, tested. Four seats, because a departure in a
// two-player game ends the game before anything can inherit anything
// (ADR 0060 Decision 5).

// pileSplitFor queues Fact or Fiction's shape: a card the CASTER
// controls asks an OPPONENT to separate the caster's cards into two
// piles. Returns the prompt's ID.
func pileSplitFor(t *testing.T, g *game.Game, caster, splitter *game.Player) uuid.UUID {
	t.Helper()
	source := battlefieldCard(g, caster, game.Card{Name: "Fact or Fiction", TypeLine: "Instant"})
	if caster.Hand == nil || len(caster.Hand.Cards) < 3 {
		t.Fatalf("setup: the caster holds %d cards, need 3", len(caster.Hand.Cards))
	}
	cards := []uuid.UUID{
		caster.Hand.Cards[0].InstanceID,
		caster.Hand.Cards[1].InstanceID,
		caster.Hand.Cards[2].InstanceID,
	}
	g.WithWriteLock(func() {
		g.QueuePileSplitForEffect(game.PileSplitPrompt{
			Splitter:      splitter.ID,
			Chooser:       caster.ID,
			Owner:         caster.ID,
			Source:        source,
			SplitQuestion: "Separate those cards into two piles",
			PickQuestion:  "Take a pile",
			Cards:         cards,
			Then:          func(*game.Game, []uuid.UUID, []uuid.UUID) error { return nil },
		})
	})
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				id = c.ID
			}
		}
	})
	if id == uuid.Nil {
		t.Fatalf("setup: the split prompt was not queued")
	}
	return id
}

// TestReassignedPromptIsEnumeratedForItsNewChooser — the inheritor is
// offered the prompt's answers and nothing else, which is exactly what
// the original chooser was offered before they left. No new case in
// choiceMoves, no new kind: one field moved.
func TestReassignedPromptIsEnumeratedForItsNewChooser(t *testing.T) {
	g := newTable(t)
	caster, splitter, next := g.Seats[0], g.Seats[1], g.Seats[2]
	pileSplitFor(t, g, caster, splitter)

	before := legal.EnumerateFor(g, splitter.ID)
	if len(before) == 0 {
		t.Fatal("setup: the splitter was offered nothing before conceding")
	}
	if n := len(legal.EnumerateFor(g, next.ID)); n != 0 {
		t.Fatalf("seat 2 was offered %d moves while somebody else owed a blocking prompt", n)
	}

	if err := g.Concede(splitter.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	after := legal.EnumerateFor(g, next.ID)
	if len(after) == 0 {
		t.Fatal("the seat that inherited the prompt was offered nothing — the #499 / #618 wedge")
	}
	for _, m := range after {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("the inheriting seat was offered %q as well as the prompt's answers", m.Label)
		}
	}
	if !hasLabel(after, "Separate those cards into two piles") {
		t.Errorf("the inherited prompt's answers are not among %v", labels(after))
	}
	// Every answer the enumerator offers has to be one the engine
	// accepts — the #544 contract, now across a change of chooser.
	dispatchAll(t, g, next.ID, after)

	// And the table is still gated on the prompt, from its new seat.
	if n := len(legal.EnumerateFor(g, caster.ID)); n != 0 {
		t.Errorf("the caster was offered %d moves while the reassigned prompt is open", n)
	}
	if n := len(legal.EnumerateFor(g, splitter.ID)); n != 0 {
		t.Errorf("the seat that left is still offered %d moves", n)
	}
}
