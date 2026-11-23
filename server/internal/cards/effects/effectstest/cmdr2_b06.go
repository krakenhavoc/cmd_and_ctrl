package effectstest

import (
	"github.com/google/uuid"
)

func SlicesContainsID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
