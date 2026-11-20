package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollowhenge Overlord — Creature — Wolf {4}{G}{G}, 4/4:
//
//	"Flash
//	 At the beginning of your upkeep, for each creature you control
//	 that's a Wolf or a Werewolf, create a 2/2 green Wolf creature
//	 token."
//
// The count is taken as the trigger resolves (the Overlord itself
// included, the tokens it makes not), and the tokens arrive together.
// Flash is PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e2418f8-2f32-4f79-8ce7-44eae8e764cb",
		Name:            "Hollowhenge Overlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Hollowhenge Overlord — a 2/2 Wolf for each Wolf or Werewolf you control",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("2/2 green Wolf"),
					Name:     "creature you control that's a Wolf or a Werewolf",
					Count: func(ctx *effects.Context) int {
						return len(effects.WolvesAndWerewolvesOf(ctx, ctx.Controller()))
					},
				})),
		},
	})
}
