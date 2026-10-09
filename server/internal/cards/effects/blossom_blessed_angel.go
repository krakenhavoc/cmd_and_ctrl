package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossom-Blessed Angel // Seed Suture — Creature — Angel Cleric {3}{W},
// 2/4 // Sorcery {G/W} (preparation card, CR 722):
//
//	"Flying, vigilance
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Seed Suture — "Put a +1/+1 counter on target creature. You gain 1
//	 life."
//
// Seed Suture is shared with Emergency Phytomedic (seedSutureResolve).
//
// No simplification.
func init() {
	const id = "cd61c26b-8d3c-4730-92f2-0d94d961bafd"
	Register(Spec{
		OracleID:        id,
		Name:            "Blossom-Blessed Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Seed Suture",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    seedSutureResolve,
	})
}
