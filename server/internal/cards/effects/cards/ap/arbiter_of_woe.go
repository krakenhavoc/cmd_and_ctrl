package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arbiter of Woe — Creature — Demon {4}{B}{B}, 5/4 (Foundations, #4465):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Flying
//	 When this creature enters, each opponent discards a card and loses
//	 2 life. You draw a card and gain 2 life."
//
// Goremand's additional sacrifice, paid while casting (CR 601.2h), so an
// Arbiter put onto the battlefield by another effect costs nothing and
// still triggers. The enters trigger is Hopeless Nightmare's discard and
// life loss, then a draw and a gain of 2, in the printed order. Each
// opponent chooses the card they discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "540693f3-985c-4a4a-945c-c957c5aad395",
		Name:            "Arbiter of Woe",
		Completeness:    effects.CompletenessFull,
		AdditionalCost:  effects.SacrificeCost("a creature", effects.Creature()),
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arbiter of Woe — each opponent discards a card and loses 2 life; you draw a card and gain 2 life",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
					effects.EachOpponentLosesLife{N: 2},
					effects.DrawCards{N: 1},
					effects.GainLife{Amount: 2},
				)),
		},
	})
}
