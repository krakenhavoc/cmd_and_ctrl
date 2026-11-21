package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolutionCastFor is who's open resolution_cast prompt, or nil.
func ResolutionCastFor(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == who {
			return c
		}
	}
	return nil
}

// declineResolutionCasts answers who's resolution_cast prompt with "cast
// no more of them".
func DeclineResolutionCasts(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	c := ResolutionCastFor(g, who)
	if c == nil {
		t.Fatal("no resolution_cast prompt to decline")
	}
	if err := g.ResolvePendingChoice(c.ID, who, nil); err != nil {
		t.Fatalf("decline the copies: %v", err)
	}
}
