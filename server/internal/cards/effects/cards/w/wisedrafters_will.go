package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wisedrafter's Will — Enchantment {U}:
//
//	"Your opponents play with their hands revealed.
//	 {U}, Sacrifice Wisedrafter's Will: Draw a card.
//	 {U}{U}, Sacrifice Wisedrafter's Will: Counter target spell."
//
// Telepathy's revealed-hands slot (HandsRevealed, ADR 0199) and two
// sacrifice abilities: a draw, and a counter whose target spell is chosen
// when the ability is activated. The sacrifice is a cost, so the enchantment
// is gone before either resolves and the hands are hidden again.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7e09daa4-a0e6-447b-b0c5-6b1bf728525f",
		Name:          "Wisedrafter's Will",
		Completeness:  effects.CompletenessFull,
		HandsRevealed: game.HandRevealOpponents,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, Sacrifice Wisedrafter's Will: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{U}{U}, Sacrifice Wisedrafter's Will: Counter target spell.",
				Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.SacrificeThis()),
				Targets: effects.TargetSpell("target spell"),
				Effect:  effects.CounterTheChosenSpell,
			},
		},
	})
}
