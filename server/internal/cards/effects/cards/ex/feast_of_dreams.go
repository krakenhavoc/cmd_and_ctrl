package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feast of Dreams — Instant {1}{B} (Journey into Nyx, #5558):
//
//	"Destroy target enchanted creature or enchantment creature."
//
// A creature with an Aura attached (Enchanted, a Role included) or an
// enchantment creature; both are re-judged as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e86870b1-4028-4111-9717-7005f86afe85",
		Name:         "Feast of Dreams",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target enchanted creature or enchantment creature", effects.Or(effects.Enchanted(), effects.Enchantment())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
