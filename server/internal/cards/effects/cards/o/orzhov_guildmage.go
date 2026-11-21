package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orzhov Guildmage — Creature — Human Wizard {W/B}{W/B}, 2/2 (Ravnica:
// Clue Edition, Ravnica Allegiance):
//
//	"{2}{W}: Target player gains 1 life.
//	 {2}{B}: Each player loses 1 life."
//
// Two independent activated abilities. The life gain targets, and a target
// player who has gone is skipped (CR 608.2b). The loss is life loss, not
// damage, for every player including the controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "877bfa3f-6e30-4cb3-831e-d56849b01334",
		Name:         "Orzhov Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Target player gains 1 life.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 1}),
			Cost:    effects.ManaCost("{2}{W}"),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.GainLife{Player: p, Amount: 1}.Apply(ctx)
			},
		}, {
			Label:   "{2}{B}: Each player loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue}, // a drain answers nothing (ADR 0142)
			Cost:    effects.ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EachPlayerLifeChanges(effects.NewContext(g, item), -1)
			},
		}},
	})
}
