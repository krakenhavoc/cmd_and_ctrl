package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mox_diamond_test.go — ADR 0098 (#1744). "If this artifact would
// enter, you may discard a land card instead. If you do, put this
// artifact onto the battlefield. If you don't, put it into its owner's
// graveyard."
//
// Each owner decision of 2026-09-30 has a test here: the prompt shape
// (1), the pause across Library of Leng with the frozen entry (3), the
// effect cause (4), "if you do" as "the card left" (5, through Rest in
// Peace, which keeps the discard real), the sandbox move (6).

const MoxDiamondOracle = "f3c5978a-70fa-431f-933b-b954bd0db0ea"

// entryChoiceOfKind returns the open entry card-choice prompt of one
// kind addressed to chooser, or nil.
func EntryChoiceOfKind(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerEntryChoice answers the open entry prompt of `kind`.
func AnswerEntryChoice(t *testing.T, g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := EntryChoiceOfKind(g, kind, chooser)
	if c == nil {
		t.Fatalf("no %s prompt addressed to %s", kind, chooser)
	}
	if err := g.ResolveEntryCardChoice(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveEntryCardChoice: %v", err)
	}
}

// castMox casts Mox Diamond from the active seat's hand and passes
// priority until it resolves or its entry asks.
func CastMox(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Mox Diamond", "Artifact", MoxDiamondOracle, nil)
	PassPriorityAroundTable(t, g)
	return id
}
