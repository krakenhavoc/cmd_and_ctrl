package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terminal Criticism — Instant {1}{B} (Reality Fracture, tracker #2795):
//
//	"Destroy target creature or planeswalker that's blue or red. You gain
//	 1 life."
//
// The colour clause is a target predicate over the permanent's effective
// colours, so a creature that stopped being blue in response is no longer
// a legal target and the spell does nothing (CR 608.2b). A multicoloured
// permanent that is blue or red passes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "213b0814-f768-4408-9cb7-6f5960bcbb7a",
		Name:         "Terminal Criticism",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature or planeswalker that's blue or red",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Or(effects.OfColor("U"), effects.OfColor("R")))),
		OnResolve: effects.SocB26DestroyTheTargetAndGainLife(1),
	})
}
