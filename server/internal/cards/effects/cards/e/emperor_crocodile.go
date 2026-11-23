package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emperor Crocodile — Creature {3}{G}, 5/5:
//
//	"When you control no other creatures, sacrifice this creature."
//
// ADR 0107 §1 (#1858). A CR 603.8 state trigger. It is asked after every
// event, so a moment with no other creature — the last one flickered and
// coming back in the same resolution — still triggers it, as CR 603.8's
// example says. A board wipe that takes the Crocodile with the others is
// one event, and the Crocodile never sees itself alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26eb7ee3-c0b6-4be0-bc80-92952f55a6f9",
		Name:         "Emperor Crocodile",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenYouControlNoOther(effects.QueryType("creature"), "Emperor Crocodile — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
