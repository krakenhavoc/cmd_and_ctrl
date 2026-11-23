package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urgent Exorcism — Instant {1}{W} (Innistrad, #5572):
//
//	"Destroy target Spirit or enchantment."
//
// One target, either kind: a permanent with the Spirit subtype (any
// card type) or an enchantment, destroyed by destroyTheTargetPermanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f87ebc6-68b0-4121-bfaf-b412b077d75b",
		Name:         "Urgent Exorcism",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target Spirit or enchantment", effects.Or(effects.HasSubtype("Spirit"), effects.Enchantment())),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
