package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_of_leng_test.go — the catalog half of #650. The engine's own
// contract for the discard event is pinned in
// game/discard_replacement_test.go; what is here is the card.

const LibraryOfLengOracle = "867def48-4be8-4056-bcf1-d6b00450b9a3"

// lengDiscard asks `who` to discard `cards` as an EFFECT's instruction
// (Mind Rot's shape), answering the pick prompt straight away, and
// returns with whatever the CR 614 window queued still open.
func LengDiscard(t *testing.T, g *game.Game, who *game.Player, cards ...uuid.UUID) {
	t.Helper()
	var promptID uuid.UUID
	g.WithWriteLock(func() {
		promptID = g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: who.ID, N: len(cards)})
	})
	if promptID == uuid.Nil {
		t.Fatal("no discard prompt was queued")
	}
	if err := g.ResolveChooseCards(promptID, who.ID, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// answerLengPrompt answers the single queued "may".
func AnswerLengPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want exactly one Library of Leng prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != game.PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", c.Kind, game.PendingChoiceOptionalReplacement)
	}
	if err := g.ResolveOptionalReplacement(c.ID, chooser, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}
