package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "bfbc12ec-ecc0-4e23-a9d1-d13babd99a46",
		Name:            "Spitting Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{R}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Spitting Drake — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
