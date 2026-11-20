package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Complete Disregard — Instant {2}{B} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Exile target creature with power 3 or less."
//
// Exile, not destroy: indestructible and regeneration don't save the
// creature. The power bound is judged at announce and again at
// resolution (CR 608.2b), so a creature pumped above 3 in response is
// not exiled. Devoid is the engine's colour-removing keyword (CR
// 702.114a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0bd842c-d518-42b3-850c-22a70f7f1436",
		Name:            "Complete Disregard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         TargetCreature("target creature with power 3 or less", PowerLE(3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
