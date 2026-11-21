package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desecration Plague — Sorcery {3}{G} (Journey into Nyx, #5558):
//
//	"Destroy target enchantment or land."
//
// Creeping Mold's destroy, without the artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "131d55d1-7429-4940-b7d2-d759aabff47b",
		Name:         "Desecration Plague",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment or land", effects.Or(effects.Enchantment(), effects.Land())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
