package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Terrible Secrets — Creature — Spirit {3}{B}, 3/4 (Kamigawa: Neon
// Dynasty, #5493):
//
//	"When this creature enters, if you control an artifact and an enchantment,
//	 you draw a card and you gain 1 life."
//
// An intervening "if" (CR 603.4): checked as it triggers and as it resolves.
// The creature is neither, so it never satisfies the clause by itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7d8784f-8d2d-4f7b-a6fb-5c841936d9e5",
		Name:         "Kami of Terrible Secrets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.NeoB08CondArtifactAndEnchantment,
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.NeoB08WhenArtifactAndEnchantment),
					"Kami of Terrible Secrets — draw a card and gain 1 life",
					effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: 1}))),
		},
	})
}
