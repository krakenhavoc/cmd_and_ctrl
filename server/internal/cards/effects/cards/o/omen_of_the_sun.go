package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omen of the Sun — Enchantment {2}{W} (Theros Beyond Death, #5514):
//
//	"Flash
//	 When this enchantment enters, create two 1/1 white Human Soldier
//	 creature tokens and you gain 2 life.
//	 {2}{W}, Sacrifice this enchantment: Scry 2."
//
// Omen of the Sea's frame: flash, an enters trigger, and a sacrifice-
// for-scry-2 activated ability, each Omen paying {2} and one mana of its
// own color. The enters trigger makes the two Human Soldiers, then the
// life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08740dda-0629-4b6f-bbcc-c8088105b5e7",
		Name:            "Omen of the Sun",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Omen of the Sun — create two 1/1 white Human Soldier creature tokens and you gain 2 life",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2}, effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
