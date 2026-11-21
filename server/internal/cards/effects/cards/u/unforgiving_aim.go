package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unforgiving Aim — Instant {2}{G} (Lorwyn Eclipsed, #4458):
//
//	"Choose one —
//	 • Destroy target creature with flying.
//	 • Destroy target enchantment.
//	 • Create a 2/2 black and green Elf creature token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be7e220e-272e-45ca-8265-8d78da2cd2a7",
		Name:         "Unforgiving Aim",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target creature with flying.", effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
			effects.Mode("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment())),
			effects.Mode("Create a 2/2 black and green Elf creature token."),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(2) {
				return effects.CreateToken{Template: effects.TokenCard("2/2 black and green Elf"), N: 1}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
