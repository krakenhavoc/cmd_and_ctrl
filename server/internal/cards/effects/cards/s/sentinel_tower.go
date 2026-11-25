package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel Tower — Artifact {4} (Battlebond):
//
//	"Whenever an instant or sorcery spell is cast during your turn, this
//	 artifact deals damage to any target equal to 1 plus the number of instant
//	 and sorcery spells cast before that spell this turn."
//
// Aria of Flame's targeting cast trigger over any player's instant or sorcery
// spell during its controller's turn. The target is chosen as the trigger goes
// on the stack (CR 603.3d); the amount is read from the turn's cast log as the
// triggering spell's own place in it, so the spells cast before it by every
// player count and it does not count itself. The Tower deals the damage
// (last-known information if it has gone, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49b86b51-e1ef-4e27-84d2-66166481e0a5",
		Name:         "Sentinel Tower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.OldB206InstantOrSorceryCastOnYourTurn,
				"Sentinel Tower — deals damage to any target equal to 1 plus the number of instant and sorcery spells cast before that spell this turn",
				effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
					return 1 + effects.OldB206InstantsAndSorceriesCastBefore(ctx.Game, ctx.Trigger().Event)
				})),
				effects.TargetAny()),
		},
	})
}
