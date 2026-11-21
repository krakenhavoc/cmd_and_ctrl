package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huddle Up — Sorcery {2}{U}:
//
//	"Assist (Another player can pay up to {2} of this spell's cost.)
//	 Two target players each draw a card."
//
// Axis of Mortality's two-player pick (WithCount(2, 2)), each picked
// player drawing one. A player who has left the game by resolution is
// an illegal target and draws nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "81776a16-bf38-450b-995b-beb29121cab2",
		Name:            "Huddle Up",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         effects.TargetPlayer("two target players").WithCount(2, 2),
		OnResolve:       func(_ *game.StackItem, ctx *effects.Context) error { return effects.Cmdr2B23EachTargetPlayerDraws(ctx) },
	})
}
