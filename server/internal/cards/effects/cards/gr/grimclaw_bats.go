package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grimclaw Bats — Creature — Bat {1}{B} (Darksteel, #5605):
//
//	"Flying
//	 {B}, Pay 1 life: This creature gets +1/+1 until end of turn."
//
// Shivan Dragon's self pump with a life payment beside the mana (CR
// 119.4: a player at 0 life can't pay it). It may be activated any number
// of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "161ebbc8-b8f9-4ac4-80f3-e290070a1e30",
		Name:            "Grimclaw Bats",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Pay 1 life: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.PayLife(1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Grimclaw Bats — +1/+1"}.Apply(ctx)
			},
		}},
	})
}
