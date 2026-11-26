package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trigon of Thought — Artifact {5} (Scars of Mirrodin, S82 OLD batch 470,
// #5577):
//
//	"This artifact enters with three charge counters on it.
//	 {U}{U}, {T}: Put a charge counter on this artifact.
//	 {2}, {T}, Remove a charge counter from this artifact: Draw a card."
//
// The Trigon shape (OldB470Trigon) with a draw.
//
// No simplification.
func init() {
	effects.Register(effects.OldB470Trigon(
		effects.Spec{OracleID: "8a179d68-c9e6-4f46-99bf-11b2a2708990", Name: "Trigon of Thought", Completeness: effects.CompletenessFull},
		"{U}{U}", "Draw a card.",
		nil, game.Purpose{Draws: 1, Answers: game.AnswerValue},
		func(g *game.Game, item *game.StackItem) error {
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
		}))
}
