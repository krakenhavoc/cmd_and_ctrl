package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burnout Bashtronaut — Creature — Goblin Warrior {R}, 1/1:
//
//	"Menace
//	 Start your engines!
//	 {2}: This creature gets +1/+0 until end of turn.
//	 Max speed — This creature has double strike."
//
// ADR 0138 (#2122). Double strike is a layer 6 grant switched on by its
// controller's speed (CR 702.178a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b122fbb-c2e1-42d1-bfc5-fcacacdfa0bc",
		Name:            "Burnout Bashtronaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", effects.StartYourEngines},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Burnout Bashtronaut — +1/+0"}.Apply(ctx)
			},
		}},
		Static: effects.MaxSpeedSelfKeywords("double strike"),
	})
}
