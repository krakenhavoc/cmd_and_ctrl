package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leadership Vacuum — Instant {2}{U} (Commander 2019, #5517):
//
//	"Target player returns each commander they control from the
//	 battlefield to the command zone.
//	 Draw a card."
//
// Every commander the target player controls as the spell resolves goes to
// its OWNER's command zone (a commander they borrowed goes home to the
// player who owns it), through Hellkite Courser's door
// (game.ReturnToCommandZoneForEffect): the move to the command zone is not
// one CR 903.9b replaces, so nothing is asked, and no tax is paid. Then the
// caster draws. A target that has left the game takes the whole spell with
// it (CR 608.2b), so nobody draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f9dd005-1648-451e-b888-69ef81331c95",
		Name:         "Leadership Vacuum",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.FirstLegalTargetPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			var commanders []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCommander && c.Controller == player {
					commanders = append(commanders, c.InstanceID)
				}
			}
			for _, id := range commanders {
				if err := ctx.Game.ReturnToCommandZoneForEffect(id); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
