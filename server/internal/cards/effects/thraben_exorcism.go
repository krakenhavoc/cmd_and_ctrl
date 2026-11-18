package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thraben Exorcism — Instant {1}{W} (Innistrad: Midnight Hunt, S82):
//
//	"Exile target Spirit, creature with disturb, or enchantment."
//
// One target clause over three kinds of permanent: a Spirit (any permanent
// with the Spirit subtype, read through the layers), a creature that has
// disturb (CR 702.146a, the keyword the permanent's front face prints; a
// disturbed permanent's back face does not have it), or an enchantment.
// Angelic Edict's exile of the still-legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b43e25d8-1c8b-4316-84f3-c4ec6f77092e",
		Name:         "Thraben Exorcism",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetPermanent("target Spirit, creature with disturb, or enchantment",
			Or(HasSubtype("Spirit"), And(Creature(), WithFlashbackOrDisturb()), Enchantment())),
		OnResolve: ResolveSteps(Instruction(exileTheFirstTargetIfLegal)),
	})
}
