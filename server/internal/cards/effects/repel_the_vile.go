package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repel the Vile — Instant {3}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Choose one —
//	 • Exile target creature with power 4 or greater.
//	 • Exile target enchantment."
//
// Each bullet's target clause is its own, so the picker offers creatures
// of power 4 or more, or enchantments, by mode.
//
// No simplification.
func init() {
	exile := func(text string, spec *game.TargetSpec) game.ModeOption {
		return ModeWithPurpose(ModeDoing(text, spec, exileTheModesTarget), game.Purpose{Answers: game.AnswerRemove})
	}
	Register(Spec{
		OracleID:     "aa108554-f1a0-45b6-aa15-597d42be7a8a",
		Name:         "Repel the Vile",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			exile("Exile target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4))),
			exile("Exile target enchantment.",
				TargetPermanent("target enchantment", Enchantment())),
		),
	})
}
