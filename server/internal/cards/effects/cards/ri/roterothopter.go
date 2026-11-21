package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roterothopter — Artifact Creature — Thopter {1}, 0/2 (Homelands, #5752):
//
//	"Flying
//	 {2}: This creature gets +1/+0 until end of turn. Activate no more
//	 than twice each turn."
//
// The limit is AtMostNActivationsPerTurn(label, 2): the per-object count
// of announcements this turn (CR 602.1b), so a third activation is never
// offered, whether or not the first two have resolved, and the count
// starts over next turn. The pump is Barrow Naughty's self boost.
//
// No simplification.
func init() {
	const label = "{2}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn."
	effects.Register(effects.Spec{
		OracleID:        "69c69d69-3565-4d9c-82c0-a96d0fb80513",
		Name:            "Roterothopter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{2}"),
			Condition: effects.AtMostNActivationsPerTurn(label, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Roterothopter — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
