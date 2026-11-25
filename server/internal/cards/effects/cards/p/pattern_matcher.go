package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pattern Matcher — Artifact Creature — Golem {4}, 3/3 (Core Set 2020, #5518):
//
//	"When this creature enters, you may search your library for a card with the
//	 same name as another creature you control, reveal it, put it into your
//	 hand, then shuffle."
//
// The names are read from the other creatures you control as the trigger
// resolves; the search may find nothing and may be declined.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5b33c1a-6e15-4a08-b834-38b0465b0cfb",
		Name:         "Pattern Matcher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pattern Matcher — you may search for a card with the same name as another creature you control", effects.OldB154PatternMatcherSearch),
		},
	})
}
