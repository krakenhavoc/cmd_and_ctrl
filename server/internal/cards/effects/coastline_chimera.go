package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coastline Chimera — Creature — Chimera {3}{U}, 1/5:
//
//	"Flying
//	 {1}{W}: This creature can block an additional creature this turn."
//
// The activation is BlockCapacityUntilEOT on itself (#1715): an ADR
// 0041 data record adding one to its block capacity until cleanup.
// Activating it twice lets it block three, because the records add up
// exactly as two "an additional creature" statics do.
func init() {
	Register(Spec{
		OracleID:        "da95382c-0536-49aa-a0db-3b52926bf559",
		Name:            "Coastline Chimera",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: This creature can block an additional creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{W}"),
			Effect:  selfBlocksAdditionalThisTurn("Coastline Chimera — can block an additional creature this turn"),
		}},
	})
}

// selfBlocksAdditionalThisTurn is the effect of "{cost}: This creature
// can block an additional creature this turn" (Coastline Chimera,
// Mounted Archers): one more attacker for the ability's source.
func selfBlocksAdditionalThisTurn(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return BlockCapacityUntilEOT{Target: item.SourceCardID, Additional: 1, Label: label}.Apply(NewContext(g, item))
	}
}
