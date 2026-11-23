package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cfc5a284-7238-4c98-9d2f-7e5c3329be7b",
		Name:         "Blood Pact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B32TargetPlayerDrawsAndLosesLife(item, ctx, 2, 2)
		},
	})
}
