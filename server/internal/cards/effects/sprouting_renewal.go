package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "674202f0-6832-4392-a118-d513162bcd03",
		Name:         "Sprouting Renewal",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Create a 2/2 green and white Elf Knight creature token with vigilance.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard("2/2 green and white Elf Knight with vigilance"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
			ModeWithPurpose(ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				func(item *game.StackItem, ctx *Context, _ int) error { return destroyTheTargetPermanent(item, ctx) }),
				game.Purpose{Answers: game.AnswerRemove}),
		),
	})
}
