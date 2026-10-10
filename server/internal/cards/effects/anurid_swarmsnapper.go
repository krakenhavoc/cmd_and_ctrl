package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anurid Swarmsnapper — Creature — Frog Beast {2}{G}, 1/4:
//
//	"Reach (This creature can block creatures with flying.)
//	 {1}{G}: This creature can block an additional creature this turn."
//
// Reach is the engine's keyword; the activation is Coastline
// Chimera's, for {1}{G} (#1715).
func init() {
	Register(Spec{
		OracleID:        "fd275ea7-9ccc-4148-9b33-392a612486dd",
		Name:            "Anurid Swarmsnapper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: This creature can block an additional creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{G}"),
			Effect:  selfBlocksAdditionalThisTurn("Anurid Swarmsnapper — can block an additional creature this turn"),
		}},
	})
}
