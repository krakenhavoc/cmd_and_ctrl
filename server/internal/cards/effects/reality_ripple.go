package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reality Ripple — Instant {1}{U}:
//
//	"Target artifact, creature, or land phases out. (While it's phased
//	 out, it's treated as though it doesn't exist. It phases in before
//	 its controller untaps during their next untap step.)"
//
// Vodalian Illusionist's spell half, and it is here for the target
// clause rather than for the effect: an artifact, a creature or a
// LAND, which is the widest phasing target printed and the one that
// proves the primitive does not care what type of permanent it is
// given. A land phased out taps for nothing, is not counted by a
// "lands you control" static, and is not a land drop when it comes
// back — all of which fall out of it not being on the battlefield
// slice, with nothing here to say so.
//
// A Reality Ripple on a blocker after blockers are declared still
// leaves the attacker blocked (CR 509.1h); on the only creature
// targeted by a removal spell it is a counterspell that gives the
// creature back. The "or turn target face-down creature face up"
// half some printings carry is not on this oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "504d5c29-7c37-4c31-8549-2a49eeef74c8",
		Name:         "Reality Ripple",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, creature, or land",
			Or(Artifact(), Creature(), Land())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PhaseOut{Targets: legalTargetCards(item, ctx.Game)}.Apply(ctx)
		},
	})
}
