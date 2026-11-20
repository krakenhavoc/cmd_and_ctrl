package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repel the Abominable — Instant {1}{W}:
//
//	"Prevent all damage that would be dealt this turn by non-Human
//	 sources."
//
// #2026's negation over a subtype, with no positive property: any
// source that is not a Human, read as it would deal damage (CR 609.7b).
// The ruling: a source with no creature type at all is a non-Human
// source, and a Human Werewolf is still a Human; a changeling has every
// creature type, so it is a Human (CR 702.73a). Combat damage or not,
// to anything.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ed198c68-c219-469f-b133-737b060cfbfc",
		Name:         "Repel the Abominable",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldAnything, Filter: exceptSubtypes("Human")}),
	})
}
