package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disembowel — Instant {X}{B} (Ravnica: City of Guilds, #5600):
//
//	"Destroy target creature with mana value X."
//
// Entrancing Melody's target (the creature's mana value must equal the X
// announced, CR 107.3a, judged at announce and again at resolution, CR
// 608.2b) with the ordinary destroy, which regeneration can replace.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6fa90258-3c9f-4723-880f-31871c0772e0",
		Name:         "Disembowel",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with mana value X").WithManaValueEqualsX(),
		OnResolve:    destroyTheTargetPermanent,
	})
}
