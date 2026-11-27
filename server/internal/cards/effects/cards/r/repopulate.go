package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Repopulate — Instant {1}{G}:
//
//	"Shuffle all creature cards from target player's graveyard into that
//	 player's library.
//	 Cycling {2}"
//
// Every creature card in the graveyard goes into its owner's library, then
// that library is shuffled, even when there were none. A target player who is
// no longer legal leaves the spell with nothing to do (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a225d39-c1f8-4ab0-98ec-dad63bf268db",
		Name:         "Repopulate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			player := item.Targets[0].ID
			for _, id := range effects.GraveyardCardIDs(ctx, player, func(c game.Card) bool { return c.IsCreature() }) {
				if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.ShuffleLibrary{Player: player}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
