package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Encircling Fissure — Instant {2}{W}:
//
//	"Prevent all combat damage that would be dealt this turn by creatures
//	 target opponent controls.
//	 Awaken 2—{4}{W}"
//
// ADR 0135 §3 (#2411), on #2026's controller filter: the shield names the
// target opponent as it resolves (game.SourceControllerPlayer), and is
// read as each creature would deal combat damage (CR 609.7b), so a
// creature that player gains control of later this turn is caught. Then
// the awaken land (CR 702.113a).
//
// No simplifications.
func init() {
	t := effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:     "39d7569e-62b0-4f37-813a-0985632c66c9",
		Name:         "Encircling Fissure",
		Completeness: effects.CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(2, "{4}{W}", t),
		},
		OnResolve: effects.AwakenAfter(2, func(_ *game.StackItem, ctx *effects.Context) error {
			p, ok := ctx.ClauseTarget(0)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			return effects.CombatShieldAgainstCreatures(game.DamageSourceFilter{
				Controller: game.SourceControllerPlayer, ControllerPlayer: p.ID,
			}).Apply(ctx)
		}),
	})
}
