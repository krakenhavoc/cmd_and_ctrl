package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quill-Blade Laureate // Twofold Intent — Creature — Human Cleric {1}{W},
// 1/1 // Sorcery {1}{W} (preparation card, CR 722; Secrets of Strixhaven,
// #4456):
//
//	"Double strike
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Twofold Intent — "Target creature gets +1/+0 and gains double strike
//	 until end of turn."
//
// Double strike is a printed keyword; the prepare spell is Interjection's
// body (socB27TargetGetsAndGains) at +1/+0 and double strike.
//
// No simplification.
func init() {
	const id = "bd9d3a70-faed-4bc6-97d8-1601b36cdb21"
	Register(Spec{
		OracleID:        id,
		Name:            "Quill-Blade Laureate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Twofold Intent",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "double strike")),
		OnResolve:    socB27TargetGetsAndGains(1, 0, "double strike", "Twofold Intent"),
	})
}
