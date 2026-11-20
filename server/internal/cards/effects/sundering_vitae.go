package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "132c8a62-00bc-4b89-b775-099d8909f200",
		Name:         "Sundering Vitae",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
