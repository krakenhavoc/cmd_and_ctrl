package game

import (
	"testing"

	"github.com/google/uuid"
)

// takeSearchPrompts answers every open search prompt by taking as many
// of its candidates as it allows, in library order. Since #3053 a search
// for a card with a stated quality always asks (CR 701.23b — the searcher
// may find nothing, so there is never "nothing to decide"), so a test
// that used to read the result straight off a one-match search answers
// the prompt first. Tests about the prompt itself answer it by hand.
func takeSearchPrompts(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var choice *PendingChoice
		g.ReadSnapshot(func() {
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == PendingChoiceSearchLibrary {
					choice = c
					return
				}
			}
		})
		if choice == nil {
			return
		}
		n := min(choice.SearchMax, len(choice.SearchCards))
		picks := append([]uuid.UUID(nil), choice.SearchCards[:n]...)
		if err := g.ResolveSearchLibrary(choice.ID, choice.Chooser, picks); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
}
