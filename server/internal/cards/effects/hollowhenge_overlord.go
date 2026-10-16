package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0e2418f8-2f32-4f79-8ce7-44eae8e764cb",
		Name:            "Hollowhenge Overlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Hollowhenge Overlord — a 2/2 Wolf for each Wolf or Werewolf you control",
				Do(CreateTokensForEach{
					Template: TokenCard("2/2 green Wolf"),
					Name:     "creature you control that's a Wolf or a Werewolf",
					Count: func(ctx *Context) int {
						return len(wolvesAndWerewolvesOf(ctx, ctx.Controller()))
					},
				})),
		},
	})
}
