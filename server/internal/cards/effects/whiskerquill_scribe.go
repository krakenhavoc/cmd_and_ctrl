package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiskerquill Scribe — Creature — Mouse Citizen {1}{R}, 2/2 (Bloomburrow,
// #5230):
//
//	"Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, you may discard a
//	 card. If you do, draw a card."
//
// The shared valiant trigger (valiant.go) with Flaring Cinder's rummage:
// the discard is chosen as the ability resolves, and a card is drawn only
// when one was discarded.
//
// No simplification.
func init() {
	const label = "Whiskerquill Scribe — Valiant: you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "cb1e1acc-3805-418b-bed1-65838ddb1d62",
		Name:         "Whiskerquill Scribe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Valiant(label, Do(Instruction(b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)))),
		},
	})
}
