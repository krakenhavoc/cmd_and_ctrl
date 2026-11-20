package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urgent Exorcism — Instant {1}{W} (Innistrad, #5572):
//
//	"Destroy target Spirit or enchantment."
//
// One target, either kind: a permanent with the Spirit subtype (any
// card type) or an enchantment, destroyed by destroyTheTargetPermanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f87ebc6-68b0-4121-bfaf-b412b077d75b",
		Name:         "Urgent Exorcism",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target Spirit or enchantment", Or(HasSubtype("Spirit"), Enchantment())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
