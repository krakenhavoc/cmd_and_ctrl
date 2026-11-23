package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atris, Oracle of Half-Truths — Creature — Human Warlock {2}{U}{B}, 3/3:
//
//	"Menace
//	 When Atris enters, target opponent looks at the top three cards
//	 of your library and separates them into a face-down pile and a
//	 face-up pile. Put one pile into your hand and the other into your
//	 graveyard."
//
// Menace is a printed keyword; the enters trigger is the face-down
// split over three cards with a targeted opponent
// (face_down_piles.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b2b00ab-f1c5-4057-9957-d7daac95a847",
		Name:            "Atris, Oracle of Half-Truths",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:       "Atris — target opponent separates the top three cards into a face-down and a face-up pile",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.OpponentSeparatesFaceDownPiles(ctx, "Atris, Oracle of Half-Truths", 3, effects.OpponentTarget(ctx), effects.FaceDownPilesToHandTail)
			},
		}},
	})
}
