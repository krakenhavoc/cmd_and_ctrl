package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grixis Battlemage — Creature — Human Wizard {2}{B}, 2/2 (Shards of Alara,
// S82 OLD batch 509, #5588):
//
//	"{U}, {T}: Draw a card, then discard a card.
//	 {R}, {T}: Target creature can't block this turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89838d28-6026-417d-9ec3-0081911291fb",
		Name:         "Grixis Battlemage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B16DrawThenDiscard(g, item, 1)
				},
			},
			{
				Label:   "{R}, {T}: Target creature can't block this turn.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
				Effect:  effects.TargetCreatureCantBlockThisTurn("Grixis Battlemage — can't block"),
			},
		},
	})
}
