package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discover_test.go — the card-side words for CR 701.57 (ADR 0099). The
// engine rules are pinned in game/discover_test.go; these pin that the
// primitive discovers for the right player and hands its continuation
// a context, and that WheneverYouDiscover watches only its controller.

func AnswerMayCastPrompt(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if IsKeywordCastPrompt(c) && c.Chooser == chooser {
			AnswerKeywordCast(t, g, c, apply)
			return
		}
	}
	t.Fatalf("no may_cast prompt outstanding")
}

func DiscoverLibraryCard(owner uuid.UUID, name, typeLine, cost string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: owner, Controller: owner}
}
