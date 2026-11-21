package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "5513a221-dfff-4e02-8037-dcff3d0381d7",
		Name:             "Petty Larceny",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		AlternativeCosts: []game.AlternativeCost{Freerunning("{1}{B}", "Assassin")},
		Targets:          TargetPlayer("target opponent", Opponent()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return gontiExileTopForPlay(ctx, t.ID, ctx.Controller(), 2)
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
			}),
		),
	})
}
