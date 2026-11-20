package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fungal Rebirth — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "b917f39a-3176-47ad-afda-08d09b873c02", Name: "Fungal Rebirth", Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error { return effects.ReturnTargetFromGraveyardToHand(ctx.Item, ctx) }), effects.Instruction(func(ctx *effects.Context) error {
			if effects.B11CreaturesDiedThisTurn(ctx.Game) == 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 2}.Apply(ctx)
		})),
	})
}
