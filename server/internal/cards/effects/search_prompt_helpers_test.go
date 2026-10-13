package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// takeSearchPrompts answers every open search prompt, for every seat, by
// taking as many of its candidates as it allows, in library order.
//
// Since #3053 a search for a card with a stated quality always asks
// (CR 701.23b — the searcher may find nothing, so there is never
// "nothing to decide"), including when exactly one card matches. A test
// that used to read the result straight off a one-match fetch therefore
// answers the prompt first, with this. A test about the prompt itself
// answers it by hand (answerSearchNamed / answerSearchFailToFind).
func takeSearchPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var open *game.PendingChoice
		g.ReadSnapshot(func() {
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
					open = c
					return
				}
			}
		})
		if open == nil {
			return
		}
		n := min(open.SearchMax, len(open.SearchCards))
		picks := append([]uuid.UUID(nil), open.SearchCards[:n]...)
		if err := g.ResolveSearchLibrary(open.ID, open.Chooser, picks); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
}
