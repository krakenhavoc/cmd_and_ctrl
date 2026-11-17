package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spitting Drake — Creature — Drake {3}{R}, 2/2 (Visions, #5752):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn. Activate only once
//	 each turn."
//
// The limit is OncePerTurnActivation (AtMostNActivationsPerTurn at n = 1):
// the per-object count of announcements this turn (CR 602.1b). The pump is
// Barrow Naughty's self boost.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +1/+0 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:        "bfbc12ec-ecc0-4e23-a9d1-d13babd99a46",
		Name:            "Spitting Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{R}"),
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Spitting Drake — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
