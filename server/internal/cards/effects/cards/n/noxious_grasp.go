package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noxious Grasp — Instant {1}{B} (Core Set 2020, #5518):
//
//	"Destroy target creature or planeswalker that's green or white. You gain 1
//	 life."
//
// Terminal Criticism's body over green or white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b17f4960-9603-4a81-8dc8-e09e8b822357",
		Name:         "Noxious Grasp",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target creature or planeswalker that's green or white",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Or(effects.OfColor("G"), effects.OfColor("W")))),
		OnResolve: effects.SocB26DestroyTheTargetAndGainLife(1),
	})
}
