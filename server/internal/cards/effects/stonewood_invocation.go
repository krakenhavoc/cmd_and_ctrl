package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stonewood Invocation — Instant {3}{G} (Time Spiral, #5596):
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Target creature gets +5/+5 and gains shroud until end of turn. (It
//	 can't be the target of spells or abilities.)"
//
// A +5/+5 and shroud (CR 702.18a) until end of turn, behind split second
// (CR 702.61a), so no one can answer it with a removal spell while it is
// on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "35f09a11-1d9b-4dfd-9695-b809c871f10d",
		Name:            "Stonewood Invocation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         TargetCreature("target creature"),
		Purpose:         withAnswers(ForTargets(BenefitUntilEOT(0, 5, 5, "shroud")), game.AnswerProtect|game.AnswerPump),
		OnResolve:       tspB1PumpAndGrantUntilEOT(5, 5, "shroud", "Stonewood Invocation"),
	})
}
