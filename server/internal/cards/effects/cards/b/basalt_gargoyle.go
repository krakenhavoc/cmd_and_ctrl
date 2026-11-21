package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basalt Gargoyle — Creature — Gargoyle, {2}{R}, 3/2:
//
//	"Flying
//	 Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {R}: This creature gets +0/+1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83a38c8d-e2f6-404b-83be-9cfce037d6e2",
		Name:            "Basalt Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Toughness: 1, Label: "Basalt Gargoyle — +0/+1"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Basalt Gargoyle", "{2}{R}")},
	})
}
