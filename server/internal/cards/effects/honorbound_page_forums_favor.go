package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honorbound Page // Forum's Favor — Creature — Cat Cleric {3}{W}, 3/3 //
// Sorcery {W} (preparation card, CR 722; Secrets of Strixhaven, #4456):
//
//	"First strike
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Forum's Favor — "Target creature gets +1/+0 and gains flying until end
//	 of turn."
//
// First strike is a printed keyword; the prepare spell is Interjection's
// body (socB27TargetGetsAndGains) at +1/+0 and flying.
//
// No simplification.
func init() {
	const id = "66ee56b9-aca8-4efa-b9b5-50717e0784a7"
	Register(Spec{
		OracleID:        id,
		Name:            "Honorbound Page",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Forum's Favor",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "flying")),
		OnResolve:    socB27TargetGetsAndGains(1, 0, "flying", "Forum's Favor"),
	})
}
