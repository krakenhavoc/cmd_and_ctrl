package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riveteers Charm — Instant {B}{R}{G} (Streets of New Capenna):
//
//	"Choose one —
//	 • Target opponent sacrifices a creature or planeswalker they control
//	   with the greatest mana value among creatures and planeswalkers they
//	   control.
//	 • Exile the top three cards of your library. Until your next end step,
//	   you may play those cards.
//	 • Exile target player's graveyard."
//
// The first bullet is Break Under Pressure's edict: the targeted opponent
// chooses among the tied greatest, and the permanent is not targeted. The
// second is one play permission over the three cards (CR 611.2c locks the
// set), closing as Ob Nixilis' window does: as this turn's end step begins
// when cast on your own turn before it, otherwise as your next turn's does.
// Play, not cast, so a land among them can be played with a land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54e1b7f3-096a-4eb0-bcce-cce036d6cbe5",
		Name:         "Riveteers Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target opponent sacrifices a creature or planeswalker they control with the greatest mana value among creatures and planeswalkers they control.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
						effects.SacrificeSpec("a creature or planeswalker with the greatest mana value",
							effects.B10GreatestManaValueCreatureOrPlaneswalkerYouControl()),
						"Riveteers Charm — sacrifice a creature or planeswalker with the greatest mana value")
					return nil
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Exile the top three cards of your library. Until your next end step, you may play those cards.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					_, err := ctx.Game.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 3, game.CastPermission{
						Duration:   ctx.Game.UntilYourNextEndStepDuration(item.Controller),
						Source:     item.SourceCardID,
						SourceName: "Riveteers Charm",
						Label:      "Play a card exiled by Riveteers Charm",
					})
					return err
				}), game.Purpose{Draws: 3}),
			effects.ModeDoing("Exile target player's graveyard.",
				effects.TargetPlayer("target player"),
				effects.ExileTheModesTargetPlayersGraveyard),
		),
	})
}
