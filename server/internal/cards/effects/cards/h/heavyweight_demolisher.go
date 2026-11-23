package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heavyweight Demolisher — Artifact Creature — Construct {7}, 8/6:
//
//	"Menace
//	 At the beginning of your upkeep, tap this creature unless you pay {3}.
//	 Unearth {6}{R}{R}"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe71ada0-ac86-4035-9b6b-4c24ae65034c",
		Name:            "Heavyweight Demolisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Heavyweight Demolisher — tap it unless you pay {3}", func(g *game.Game, item *game.StackItem) error {
				return effects.UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{3}",
					Question: "Heavyweight Demolisher — pay {3}, or tap it?",
					Else:     demolisherTapIt,
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{6}{R}{R}")},
	})
}

// demolisherTapIt is the "or else": tap the source if it is still the
// same permanent. An on-disk key.
var demolisherTapIt = effects.PayUnlessElse("pay-unless/heavyweight-demolisher-tap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.TapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
	})
