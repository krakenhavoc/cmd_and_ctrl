package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sprouting Renewal — Sorcery {2}{G} (Ravnica Remastered, #5948):
//
//	"Convoke
//	 Choose one —
//	 • Create a 2/2 green and white Elf Knight creature token with vigilance.
//	 • Destroy target artifact or enchantment."
//
// Convoke is Spec.TapCost (Artistic Refusal). The token is the table's
// 2/2 green and white Elf Knight row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "674202f0-6832-4392-a118-d513162bcd03",
		Name:         "Sprouting Renewal",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Create a 2/2 green and white Elf Knight creature token with vigilance.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("2/2 green and white Elf Knight with vigilance"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
			effects.ModeWithPurpose(effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DestroyTheTargetPermanent(item, ctx)
				}),
				game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}
