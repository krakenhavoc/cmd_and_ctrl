package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Congregate — Instant {3}{W} (Urza's Saga, #5624):
//
//	"Target player gains 2 life for each creature on the battlefield."
//
// Fog of War's count (every creature on the battlefield, whoever controls
// it), read once as the spell resolves (CR 608.2h), doubled, and given to
// the target player if they are still a legal target (CR 608.2b). The
// amount is counted at resolution, so the Purpose declares none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d139294c-36bb-4a1e-800d-50c4a4ae9877",
		Name:         "Congregate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return effects.GainLife{Player: t.ID, Amount: 2 * len(ctx.CreatureIDs())}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
