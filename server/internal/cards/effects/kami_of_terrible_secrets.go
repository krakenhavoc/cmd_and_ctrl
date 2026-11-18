package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d7d8784f-8d2d-4f7b-a6fb-5c841936d9e5",
		Name:         "Kami of Terrible Secrets",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(neoB08CondArtifactAndEnchantment,
				On(game.EventETB, AllOf(Self, neoB08WhenArtifactAndEnchantment),
					"Kami of Terrible Secrets — draw a card and gain 1 life",
					Do(DrawCards{N: 1}, GainLife{Amount: 1}))),
		},
	})
}
