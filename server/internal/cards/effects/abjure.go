package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abjure — Instant {U} (Weatherlight, #5629):
//
//	"As an additional cost to cast this spell, sacrifice a blue permanent.
//	 Counter target spell."
//
// The sacrifice is paid as the spell is cast (CR 601.2f), so a blue
// permanent the caster controls must be there to cast it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "dd97e457-07eb-4918-b780-ae7ff42a37e7",
		Name:           "Abjure",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: SacrificeCost("a blue permanent", OfColor("U")),
		Targets:        TargetSpell("target spell"),
		OnResolve:      counterTheTargetSpell,
	})
}
