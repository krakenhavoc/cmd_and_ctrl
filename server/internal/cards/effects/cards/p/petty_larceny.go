package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Petty Larceny — Sorcery {3}{B}:
//
//	"Freerunning {1}{B} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Look at the top two cards of target opponent's library and exile
//	 those cards face down. You may play those cards for as long as they
//	 remain exiled, and mana of any type can be spent to cast them. Create
//	 a Treasure token."
//
// Outrageous Robbery's exile on two cards: face down, playable by the
// caster alone while they stay in exile, lands included, spending mana of
// any type (CR 609.4b). The Treasure comes after the exile, once any
// prompt the exile asks is answered (CR 608.2c). With the target gone as
// the spell resolves it does nothing at all (CR 608.2b). Freerunning is
// Eagle Vision's alternative cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5513a221-dfff-4e02-8037-dcff3d0381d7",
		Name:             "Petty Larceny",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{1}{B}", "Assassin")},
		Targets:          effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return effects.GontiExileTopForPlay(ctx, t.ID, ctx.Controller(), 2)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}),
		),
	})
}
