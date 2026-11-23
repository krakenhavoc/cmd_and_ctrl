package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoke Spirits' Aid — Sorcery {X}{R} (Kamigawa: Neon Dynasty Commander, S82
// #5493):
//
//	"For each of up to X target creatures, create a red Aura enchantment
//	 token named Smoke Blessing attached to that creature. Those tokens
//	 have enchant creature and "When enchanted creature dies, it deals 1
//	 damage to its controller and you create a Treasure token.""
//
// "Up to X targets" is CountFromX with UpToX; each target still legal as
// the spell resolves gets its own Smoke Blessing, created under the
// caster's control and attached as it enters (Scriv's Contract is the
// precedent). The token's trigger is the Aura's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fef58ff-4743-4065-9df0-c5edf9aa2502",
		Name:         "Smoke Spirits' Aid",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.MarchOfSwirlingMistTargets(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.X() <= 0 {
				return nil // X is the number of targets: with X = 0 there are none
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if _, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.SmokeBlessingToken()), 1, game.TokenEntryOptions{
					AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: t.ID},
				}); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
