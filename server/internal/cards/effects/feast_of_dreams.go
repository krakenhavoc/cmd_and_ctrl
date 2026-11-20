package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feast of Dreams — Instant {1}{B} (Journey into Nyx, #5558):
//
//	"Destroy target enchanted creature or enchantment creature."
//
// A creature with an Aura attached (Enchanted, a Role included) or an
// enchantment creature; both are re-judged as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e86870b1-4028-4111-9717-7005f86afe85",
		Name:         "Feast of Dreams",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target enchanted creature or enchantment creature", Or(Enchanted(), Enchantment())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
