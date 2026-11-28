package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Allied Strategies — Sorcery {4}{U} (Planeshift, S82 OLD b642, #5615):
//
//	"Domain — Target player draws a card for each basic land type among lands
//	 they control."
//
// "They" is the target: the count is DomainCount of the targeted player's lands,
// read as the spell resolves (CR 608.2c), and that player draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5ce2ccd-8c78-4749-8854-cb5fa5f4738e",
		Name:         "Allied Strategies",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p, n, ok := effects.OldB642TargetPlayerDomain(ctx)
			if !ok {
				return nil
			}
			return effects.DrawCards{Player: p, N: n}.Apply(ctx)
		},
	})
}
