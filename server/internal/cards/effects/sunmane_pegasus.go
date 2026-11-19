package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunmane Pegasus — Creature — Pegasus {3}{W}, 2/3 (Foundations Jumpstart):
//
//	"Flying
//	 {1}{W}: This creature gains vigilance and lifelink until end of turn."
//
// One ability granting both keywords to this creature only, until end of
// turn (CR 611.2a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f7e7c59-e0a9-4b32-96b9-b1abf52f5ef7",
		Name:            "Sunmane Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: This creature gains vigilance and lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, "Sunmane Pegasus — vigilance and lifelink",
					game.AddKeywordsMod("vigilance", "lifelink"))
			},
		}},
	})
}
