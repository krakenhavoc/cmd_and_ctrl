package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crush Contraband — Instant {3}{W}:
//
//	"Choose one or both —
//	 • Exile target artifact.
//	 • Exile target enchantment."
//
// Against All Odds' "one or both" shape (ChooseN 1..2) over the shared
// ExileTheModesTarget bullet body: the bullets resolve in printed order (CR
// 608.2c) and each reads its own target, skipping it if it is no longer
// legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecbc234d-e1d3-4bcc-81d4-fffc2c7ea16a",
		Name:         "Crush Contraband",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Exile target artifact.", TargetPermanent("target artifact", Artifact()), ExileTheModesTarget),
			ModeDoing("Exile target enchantment.", TargetPermanent("target enchantment", Enchantment()), ExileTheModesTarget),
		),
	})
}
