package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mounted Archers — Creature — Human Soldier Archer {3}{W}, 2/3:
//
//	"Reach (This creature can block creatures with flying.)
//	 {W}: This creature can block an additional creature this turn."
//
// Reach is the engine's keyword; the activation is Coastline
// Chimera's, for {W} (#1715).
func init() {
	Register(Spec{
		OracleID:        "22517690-3ec1-49c9-951f-93eb3114423a",
		Name:            "Mounted Archers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature can block an additional creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{W}"),
			Effect:  selfBlocksAdditionalThisTurn("Mounted Archers — can block an additional creature this turn"),
		}},
	})
}
