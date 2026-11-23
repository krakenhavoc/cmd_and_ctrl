package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marang River Regent // Coil and Catch — Tarkir: Dragonstorm (#4463), an
// omen card (CR 720, #5175):
//
//	Marang River Regent — Creature — Dragon {4}{U}{U}, 6/7
//	  "Flying
//	   When this creature enters, return up to two other target nonland
//	   permanents to their owners' hands."
//	Coil and Catch — Instant — Omen {3}{U}
//	  "Draw three cards, then discard a card. (Then shuffle this card
//	   into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). "Other" is object
// identity (Another), so the Regent cannot return itself. Coil and Catch
// is Ancestral Reminiscence's draw-then-discard.
//
// No simplification.

const MarangRiverRegentOracleID = "099b8d2c-b482-48dd-b6bc-dedf546d66c5"

func init() {
	effects.Register(effects.Spec{
		OracleID:        MarangRiverRegentOracleID,
		Name:            "Marang River Regent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Marang River Regent — return up to two other target nonland permanents to their owners' hands",
				effects.BounceEachLegalTarget),
				effects.Another(effects.TargetPermanent("up to two other target nonland permanents", effects.Nonland()).WithCount(0, 2))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     MarangRiverRegentOracleID + "#1",
		Name:         "Coil and Catch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
