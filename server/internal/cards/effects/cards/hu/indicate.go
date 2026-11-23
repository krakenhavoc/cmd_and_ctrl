package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indicate — Sorcery {0} (Mystery Booster 2 playtest card, #5948):
//
//	"Target permanent."
//
// The whole text is a target and no instruction: it is cast on a permanent,
// goes through the stack, and does nothing as it resolves. A permanent that
// has become an illegal target makes the spell fizzle (CR 608.2b), which is
// equally invisible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67876214-3777-41b1-935e-75dd5075fa53",
		Name:         "Indicate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
	})
}
