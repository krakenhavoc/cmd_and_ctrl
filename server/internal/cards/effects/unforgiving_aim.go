package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unforgiving Aim — Instant {2}{G} (Lorwyn Eclipsed, #4458):
//
//	"Choose one —
//	 • Destroy target creature with flying.
//	 • Destroy target enchantment.
//	 • Create a 2/2 black and green Elf creature token."
func init() {
	Register(Spec{
		OracleID:     "be7e220e-272e-45ca-8265-8d78da2cd2a7",
		Name:         "Unforgiving Aim",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Modes: ChooseOne(
			Mode("Destroy target creature with flying.", TargetCreature("target creature with flying", HasKeyword("flying"))),
			Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Create a 2/2 black and green Elf creature token."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(2) {
				return CreateToken{Template: TokenCard("2/2 black and green Elf"), N: 1}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
