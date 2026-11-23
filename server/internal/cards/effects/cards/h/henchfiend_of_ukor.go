package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Henchfiend of Ukor — Creature — Ogre, {3}{R}, 3/2:
//
//	"Haste
//	 Echo {1}{B} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {B/R}: This creature gets +1/+0 until end of turn."
//
// The pump's hybrid {B/R} is paid with either colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "91b5688d-4dd6-477b-998f-bdc7612de37a",
		Name:            "Henchfiend of Ukor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B/R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B/R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Henchfiend of Ukor — +1/+0"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Henchfiend of Ukor", "{1}{B}")},
	})
}
