package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trigon of Infestation — Artifact {4} (Scars of Mirrodin, S82 OLD batch 470,
// #5577):
//
//	"This artifact enters with three charge counters on it.
//	 {G}{G}, {T}: Put a charge counter on this artifact.
//	 {2}, {T}, Remove a charge counter from this artifact: Create a 1/1 green
//	 Phyrexian Insect creature token with infect."
//
// The Trigon shape (OldB470Trigon) making the 1/1 Insect the catalog already
// has (Contagion Clasp's).
//
// No simplification.
func init() {
	effects.Register(effects.OldB470Trigon(
		effects.Spec{OracleID: "6af2b898-7ba0-4641-8234-5171397cf602", Name: "Trigon of Infestation", Completeness: effects.CompletenessFull},
		"{G}{G}", "Create a 1/1 green Phyrexian Insect creature token with infect.",
		nil, game.Purpose{Tokens: 1, Answers: game.AnswerValue},
		func(g *game.Game, item *game.StackItem) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Phyrexian Insect with infect"), N: 1}.Apply(effects.NewContext(g, item))
		}))
}
