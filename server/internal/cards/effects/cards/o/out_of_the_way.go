package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Out of the Way — Instant {3}{U} (Streets of New Capenna, #5491):
//
//	"This spell costs {2} less to cast if it targets a green permanent.
//	 Return target nonland permanent an opponent controls to its owner's
//	 hand.
//	 Draw a card."
//
// The reduction reads the announced target (Fantastic Bounce's shape,
// CostsLessIfItTargets). The draw waits behind the bounce
// (bounceTheTargetThenDraw), so a replacement or an order prompt on the
// return is answered first. The draw is the spell's own and happens only
// if the spell resolves, which needs its target still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9300874a-32a9-4869-b3e1-1b996403df4e",
		Name:         "Out of the Way",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target nonland permanent an opponent controls",
			effects.And(effects.Nonland(), effects.OpponentControls())),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a green permanent.",
				effects.OfColor("G")),
		},
		OnResolve: effects.BounceTheTargetThenDraw,
	})
}
