package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morbid Plunder — Sorcery {1}{B}{B} (Mirrodin Besieged,
// #5576):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand."
//
// Macabre Reconstruction's return without its discount: up to two creature
// cards from your own graveyard, each returned if still there as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1bb7009b-aeb0-4545-be4c-c8c5d723bcf5",
		Name:         "Morbid Plunder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
