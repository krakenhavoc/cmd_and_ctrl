package game

import (
	"sync"

	"github.com/google/uuid"
)

// census_s11c_test.go — #4361 S11c: the closure doors ExileCards-,
// BounceCardsToHand-, TuckCardsToLibrary- and
// CounterSpellToLibraryThenForEffect are deleted, so
// the tests that read what such a continuation was told run a registered
// probe tail instead.

var (
	toldProbeMu  sync.Mutex
	toldProbeLog = map[*Game][][]uuid.UUID{}
	// toldProbe records, per game, every run of the tail and the IDs it
	// was told.
	toldProbe = RegisterTail(testEffectKeyPrefix+"s11c-told-probe", func(g *Game, o TailOutcome) error {
		toldProbeMu.Lock()
		defer toldProbeMu.Unlock()
		toldProbeLog[g] = append(toldProbeLog[g], copyUUIDs(o.IDs))
		return nil
	})
)

// toldOf is every run of toldProbe on g, in order, with what it was told.
func toldOf(g *Game) [][]uuid.UUID {
	toldProbeMu.Lock()
	defer toldProbeMu.Unlock()
	return append([][]uuid.UUID(nil), toldProbeLog[g]...)
}
