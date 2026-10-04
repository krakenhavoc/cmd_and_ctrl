package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evernight Shade — Creature — Shade {3}{B}, 1/1:
//
//	"{B}: This creature gets +1/+1 until end of turn.
//	 Undying"
//
// Undying is PrintedKeywords (#2075). A pump does not put a +1/+1
// counter on it, so a pumped Shade still returns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9af1b6c4-295f-41d2-b3a0-0863798330d4",
		Name:            "Evernight Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Activated: []ActivatedAbility{{
			Label:  "{B}: This creature gets +1/+1 until end of turn.",
			Cost:   ManaCost("{B}"),
			Effect: thisGetsUntilEndOfTurn(1, 1, "Evernight Shade — +1/+1 until end of turn"),
		}},
	})
}
