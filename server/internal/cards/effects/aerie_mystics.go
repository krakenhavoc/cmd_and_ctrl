package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerie Mystics — Creature — Bird Wizard {4}{W}, 3/3 (Conflux, #5587):
//
//	"Flying
//	 {1}{G}{U}: Creatures you control gain shroud until end of turn.
//	 (They can't be the targets of spells or abilities.)"
//
// Blinding Fog's group grant at shroud: the creatures you control as the
// ability resolves gain it (CR 611.2c), until the cleanup step (CR
// 514.2). A creature that arrives afterwards does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12134f7d-433a-416a-b668-c1a21984c94b",
		Name:            "Aerie Mystics",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{U}: Creatures you control gain shroud until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{1}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"shroud"},
					Label:    "Aerie Mystics — shroud",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
