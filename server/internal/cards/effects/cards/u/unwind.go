package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unwind — Instant {2}{U} (EDHREC rank 1062):
//
//	"Counter target noncreature spell. Untap up to three lands."
//
// Rewind's little sibling: Negate's clause with a three-land refund.
// The noncreature gate is the target predicate, checked at announce
// and again at resolution.
//
// "Untap up to three lands" is a resolution-time choice among every
// land at the table on paper. UntapUpToLands is that choice: a
// prompt over every tapped land at the table, any controller's,
// queued after the counter resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e38b8ecb-e7ae-474d-b6a4-29cc1aa8ccd9",
		Name:         "Unwind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve:    effects.B09CounterThenUntapLands(3, "Unwind — untap up to three lands"),
	})
}
