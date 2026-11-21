package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// putInLibraryChoiceFor is the newest open put_in_library prompt owed
// by `chooser`, or nil.
func PutInLibraryChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePutInLibrary && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// countEvents counts every event of `kind` in the log.
func CountEvents(g *game.Game, kind game.EventKind) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

// libraryBottomIDs is the bottom n cards of a library, top-first — so
// the LAST entry is the bottom card.
func LibraryBottomIDs(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := n - 1; i >= 0; i-- {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}

func SameIDs(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
