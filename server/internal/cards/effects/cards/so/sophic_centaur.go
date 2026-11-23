package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sophic Centaur — Creature — Centaur Spellshaper {3}{G}, 1/1 (Planar Chaos, #5595):
//
//	"{2}{G}{G}, {T}, Discard a card: You gain 2 life for each card in your
//	 hand."
//
// Gerrard's Wisdom on a spellshaper: the hand is counted as the ability
// resolves, after the discard that paid for it (CR 602.2b, 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e8e443b-2102-4d91-b50c-1679dc893ef7",
		Name:         "Sophic Centaur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{G}, {T}, Discard a card: You gain 2 life for each card in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}{G}"), effects.TapCost(), effects.DiscardACard()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: 2 * effects.B14HandSize(ctx.Game, ctx.Controller())}.Apply(ctx)
			})),
		}},
	})
}
