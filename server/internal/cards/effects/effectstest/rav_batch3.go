package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_batch3.go — Ravnica batch 3 helpers shared across test packages.

// RavB3LibraryTop is the top n cards of a library, top-first.
func RavB3LibraryTop(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := len(p.Library.Cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}

// RavB3SameOrder reports whether two ID lists match in order.
func RavB3SameOrder(got, want []uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
