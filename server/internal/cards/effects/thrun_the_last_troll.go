package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrun, the Last Troll — Legendary Creature — Troll Shaman {2}{G}{G},
// 4/4:
//
//	"This spell can't be countered.
//	 Hexproof (This creature can't be the target of spells or
//	 abilities your opponents control.)
//	 {1}{G}: Regenerate Thrun."
//
// Uncounterable through Spec.CantBeCountered, hexproof through
// PrintedKeywords, and the regeneration is Regenerate on the
// ability's own source, as for Albino Troll (CR 701.19).
//
// Not the other Thrun: Thrun, Breaker of Silence is a different
// oracle card (trample, a "nongreen" targeting restriction, and
// indestructible on its controller's turn) and is not built here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1149e5ac-554a-41b1-84ae-bac42579c1aa",
		Name:            "Thrun, the Last Troll",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"hexproof"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate Thrun.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
