package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Withering — Instant {4}{B}{B} (Time Spiral, #5596):
//
//	"Destroy target nonblack creature.
//	 Madness {B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Doom Blade at instant speed for six, or for {B} through madness (CR
// 702.35a): discarded, it goes to exile and may be cast from there for its
// madness cost, which the engine grows from the one string.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac2720d2-b010-4fa9-8c94-b13680b75aac",
		Name:         "Dark Withering",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Madness:      "{B}",
		OnResolve:    destroyTheTargetPermanent,
	})
}
