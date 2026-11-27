package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Search for Survivors — Sorcery {2}{R} (Prophecy, S82 OLD batch 673, #5617):
//
//	"Reorder your graveyard at random. An opponent chooses a card at random in
//	 your graveyard. If it's a creature card, put it onto the battlefield.
//	 Otherwise, exile it."
//
// The graveyard is shuffled in place (game.ReorderGraveyardAtRandomForEffect),
// then one card is drawn at random from the controller's graveyard from the
// game's keyed random stream; which opponent "chooses" cannot matter for a
// random draw. A creature card goes onto the battlefield under its owner's
// control (it is the controller's own graveyard); any other card is exiled. An
// empty graveyard does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02fd4da6-6c6e-4143-a915-495c668de148",
		Name:         "Search for Survivors",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.ReorderGraveyardAtRandomForEffect(ctx.Controller())
			picked := effects.RandomPick(ctx, effects.GraveyardIDs(ctx.Game, ctx.Controller(), nil), 1)
			if len(picked) == 0 {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(picked[0])
			if !ok {
				return nil
			}
			if c.IsCreature() {
				return effects.ReturnFromGraveyard{Target: c.InstanceID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
			}
			return effects.ExileTarget{Target: c.InstanceID}.Apply(ctx)
		},
	})
}
