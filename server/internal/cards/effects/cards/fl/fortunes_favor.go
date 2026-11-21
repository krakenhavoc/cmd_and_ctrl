package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortune's Favor — Instant {3}{U}:
//
//	"Target opponent looks at the top four cards of your library and
//	 separates them into a face-down pile and a face-up pile. Put one
//	 pile into your hand and the other into your graveyard."
//
// Sauron's Ransom's split with a target instead of a choice
// (face_down_piles.go). A target that is no longer legal on
// resolution fizzles the whole spell, so nothing is looked at.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9baff93e-4ef9-404d-884b-f651a11f3742",
		Name:         "Fortune's Favor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OpponentSeparatesFaceDownPiles(ctx, "Fortune's Favor", 4, effects.OpponentTarget(ctx), effects.FaceDownPilesToHandTail)
		},
	})
}
