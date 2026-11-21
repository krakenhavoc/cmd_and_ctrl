package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sundering Vitae — Instant {2}{G} (Ravnica: City of Guilds, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Destroy target artifact or enchantment."
//
// Convoke is the engine's tap cost (CR 702.51); the body is Disenchant's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "132c8a62-00bc-4b89-b775-099d8909f200",
		Name:         "Sundering Vitae",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
