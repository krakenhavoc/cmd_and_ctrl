package effectstest

import (
	"github.com/google/uuid"
)

func SameIDSet(got []uuid.UUID, want ...uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[uuid.UUID]bool, len(got))
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			return false
		}
	}
	return true
}
