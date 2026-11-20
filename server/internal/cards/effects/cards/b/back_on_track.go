package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Back on Track — Sorcery {4}{B}:
//
//	"Return target creature or Vehicle card from your graveyard to the
//	 battlefield. Create a 1/1 colorless Pilot creature token with
//	 "This token saddles Mounts and crews Vehicles as though its power
//	 were 2 greater.""
//
// One target, so a spell whose only target is gone at resolution does
// nothing at all, token included (CR 608.2b); the engine fizzles it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaf46eaf-2e01-46ac-b79d-6b70739d11c2",
		Name:         "Back on Track",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature or Vehicle card in your graveyard",
			effects.Or(effects.Creature(), effects.HasSubtype("Vehicle")), effects.YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.ReanimateSingleTarget(ctx, item.Controller); !ok {
				return nil
			}
			return effects.CreateToken{Template: effects.PilotToken(), N: 1}.Apply(ctx)
		},
	})
}
