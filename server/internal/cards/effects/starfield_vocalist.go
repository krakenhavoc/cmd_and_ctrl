package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfield Vocalist — Creature — Human Bard {3}{U}, 3/4 (Edge of
// Eternities):
//
//	"If a permanent entering the battlefield causes a triggered ability
//	 of a permanent you control to trigger, that ability triggers an
//	 additional time.
//	 Warp {1}{U}"
//
// Virtue of Knowledge's doubler (DoublesEntering with no filter on the
// entering permanent; the trigger's source must be a permanent its
// controller shares with the Vocalist's). Replacement effects and "as
// this enters" clauses are not triggers and are untouched; two Vocalists
// make three instances, not four. The Vocalist's own entry counts, and
// so does a permanent that enters with it. Warp is CR 702.185.
//
// No simplification.
func init() {
	doubler := DoublesEntering(nil)
	doubler.Label = "Starfield Vocalist"
	doubler.RowLabel = "Abilities that trigger when a permanent enters trigger an additional time."
	Register(Spec{
		OracleID:         "4add809e-9caa-45ae-bb0d-4760a360b8e2",
		Name:             "Starfield Vocalist",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{U}")},
		TriggerDoublers:  []game.TriggerDoubler{doubler},
	})
}
