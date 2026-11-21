package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Blossom-Blessed Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Seed Suture",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SeedSutureResolve,
	})
}
