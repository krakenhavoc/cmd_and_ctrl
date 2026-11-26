package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haphazard Bombardment — Enchantment {5}{R} (Dominaria):
//
//	"When this enchantment enters, choose four nonenchantment permanents you
//	 don't control and put an aim counter on each of them.
//	 At the beginning of your end step, if two or more permanents you don't
//	 control have an aim counter on them, destroy one of those permanents at
//	 random."
//
// The enters trigger does not target (a hexproof permanent can be chosen): you
// pick exactly four nonenchantment permanents among those your opponents
// control (every one when there are four or fewer), as it resolves, and each
// gets an aim counter. The end-step trigger is an intervening "if" (CR 603.4)
// over the permanents you don't control with an aim counter, any of them, whoever
// put it there; as it resolves it counts again and destroys one chosen at
// random (indestructible survives).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc1eb1c9-b22b-41a1-8b13-94a0e758db0f",
		Name:         "Haphazard Bombardment",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Haphazard Bombardment — choose four nonenchantment permanents you don't control and put an aim counter on each", effects.OldB209ChooseAimTargets),
			effects.InterveningIfYou("two or more permanents you don't control have an aim counter on them",
				func(g *game.Game, you uuid.UUID) bool { return len(effects.OldB209AimedPermanents(g, you)) >= 2 },
				effects.AtYourEndStep("Haphazard Bombardment — destroy one of those permanents at random", effects.OldB209DestroyOneAimedAtRandom)),
		},
	})
}
