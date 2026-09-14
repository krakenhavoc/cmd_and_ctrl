package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Pact — Instant {2}{B} (EDHREC rank 3387):
//
//	"Target player draws two cards and loses 2 life."
//
// Night's Whisper at instant speed, pointed at anyone. The draws
// come first, then the loss — a loss, not damage, so no prevention
// shield sees it and the player's own "whenever you lose life"
// payoffs do. A target that became illegal in response fizzles the
// spell, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfc5a284-7238-4c98-9d2f-7e5c3329be7b",
		Name:         "Blood Pact",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b32TargetPlayerDrawsAndLosesLife(item, ctx, 2, 2)
		},
	})
}
