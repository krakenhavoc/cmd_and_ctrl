package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ant-Man's Army — Creature — Insect {2}{G}, 3/2 (Marvel Super Heroes,
// #4455):
//
//	"When this creature enters, create a Food token or a Treasure token."
//
// The choice between the two tokens is made as the trigger resolves
// (CR 608.2c), the shape Tireless Provisioner's landfall already uses:
// createFoodOrTreasure, Treasure first, then Food. The step declares the
// footprint of the token it makes (ADR 0018's #3013 amendment): Food and
// Treasure are both inert templates with no trigger, static or replacement
// of their own, and enter with the same events, so one declaration is
// exact for either answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34c0f75d-10a9-428d-b202-3defb3949090",
		Name:         "Ant-Man's Army",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ant-Man's Army — create a Food token or a Treasure token",
				effects.Do(FoodOrTreasureStep{Question: "Ant-Man's Army — create a Food token or a Treasure token"})),
		},
	})
}

// foodOrTreasureStep is createFoodOrTreasure as a Do step that declares
// the footprint of a created token.
type FoodOrTreasureStep struct{ Question string }

func (s FoodOrTreasureStep) Apply(ctx *effects.Context) error {
	return effects.CreateFoodOrTreasure(s.Question)(ctx.Game, ctx.Item)
}

func (FoodOrTreasureStep) Footprint() (game.FootprintStep, bool) {
	return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Footprint()
}
