package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desecration Plague — Sorcery {3}{G} (Journey into Nyx, #5558):
//
//	"Destroy target enchantment or land."
//
// Creeping Mold's destroy, without the artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "131d55d1-7429-4940-b7d2-d759aabff47b",
		Name:         "Desecration Plague",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment or land", Or(Enchantment(), Land())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
