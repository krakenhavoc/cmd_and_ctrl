package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Break Open — Instant {1}{R} (Onslaught):
//
//	"Turn target face-down creature an opponent controls face up."
//
// Hauntwoods Shrieker's free turn (TurnFaceUp, CR 708.8): no morph cost
// is paid and no megamorph counter is put on it, it stays the same
// object, and its "when this is turned face up" abilities trigger. A
// manifested or cloaked card that is not a creature card stays face down,
// and a creature turned face up in response is no longer a legal target
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1fcec3b2-47ba-4c56-9ce1-a4807e8ad262",
		Name:         "Break Open",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target face-down creature an opponent controls", FaceDown(), OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return TurnFaceUp{Targets: legalTargetCards(item, ctx.Game)}.Apply(ctx)
		},
	})
}
