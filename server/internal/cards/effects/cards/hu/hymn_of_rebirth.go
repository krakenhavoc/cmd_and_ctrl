package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hymn of Rebirth — Sorcery {3}{G}{W} (Ice Age, S82 OLD batch 851, #5636):
//
//	"Put target creature card from a graveyard onto the battlefield under your
//	 control."
//
// Ashiok's "put target creature card from a graveyard onto the battlefield
// under your control": any player's graveyard, entering under the caster. A
// target that has left by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f06fca68-593b-4fbd-a5a9-618060311172",
		Name:         "Hymn of Rebirth",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Reanimates: true}),
		Targets:      effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82Mom16PutTargetOntoTheBattlefield(ctx.Game, item)
		},
	})
}
