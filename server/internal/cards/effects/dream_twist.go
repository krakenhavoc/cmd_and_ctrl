package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dream Twist — Instant {U} (Innistrad, #5572):
//
//	"Target player mills three cards.
//	 Flashback {1}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Dampen Thought's TargetPlayerMills at three with the shared Flashback
// ({1}{U}).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d569d969-ce39-4314-b5f0-76b45f5c4c7f",
		Name:             "Dream Twist",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return TargetPlayerMills(3)(ctx.Game, item)
		},
	})
}
