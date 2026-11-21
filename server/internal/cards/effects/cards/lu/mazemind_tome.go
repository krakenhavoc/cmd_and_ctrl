package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mazemind Tome — Artifact — Book {2}:
//
//	"{T}, Put a page counter on this artifact: Scry 1.
//	 {2}, {T}, Put a page counter on this artifact: Draw a card.
//	 When there are four or more page counters on this artifact, exile
//	 it. If you do, you gain 4 life."
//
// ADR 0107 §1 (#1858). Putting the page counter on is part of each
// row's COST (CR 602.1a), so it is there before the ability resolves and
// the fourth one triggers the exile at once, with the draw or the scry
// still on the stack under it. The exile is a CR 603.8 state trigger, and
// the 4 life comes only if the Tome really reached exile.
//
// No simplification.
func init() {
	const page = "page"
	effects.Register(effects.Spec{
		OracleID:     "800fb917-8898-4eba-9f94-aec6e9d75236",
		Name:         "Mazemind Tome",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Put a page counter on this artifact: Scry 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
				Cost:    effects.Plus(effects.TapCost(), effects.AddCounterToThis(page, 1)),
				Effect:  effects.Do(effects.Scry{N: 1}),
			},
			{
				Label:   "{2}, {T}, Put a page counter on this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.AddCounterToThis(page, 1)),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisHasAtLeast(page, 4, "Mazemind Tome — exile it and gain 4 life",
				effects.ExileThisThen("then/mazemind-tome-gain-life", func(ctx *effects.Context) error {
					return effects.GainLife{Amount: 4}.Apply(ctx)
				})),
		},
	})
}
