package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmerbell — Creature — Elemental Jellyfish {1}{U}, 1/3
// (Foundations Jumpstart):
//
//	"Flying
//	 {1}{U}: Untap this creature."
//
// Flying is a printed keyword; the untap is UntapTarget on the source,
// which does nothing once the source has left (CR 400.7). No
// simplification.
func init() {
	Register(Spec{
		OracleID:        "1a8b7792-e105-40c8-b1b4-ab1a8f3731d8",
		Name:            "Glimmerbell",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			UntapSelf: true,
			Label:     "{1}{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return UntapTarget{Target: ctx.Source()}.Apply(ctx)
			},
		}},
	})
}
