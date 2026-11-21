package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amethyst Dragon // Explosive Crystal — an adventure card (Commander Legends:
// Battle for Baldur's Gate):
//
//	Amethyst Dragon {4}{R}{R}, Creature — Dragon 4/4:
//	"Flying, haste"
//
//	Explosive Crystal {4}{R}, Sorcery — Adventure:
//	"Explosive Crystal deals 4 damage divided as you choose among any number of
//	 targets. (Then exile this card. You may cast the creature later from exile.)"
//
// The adventure layout is Brazen Borrower's (CR 715.3d). The damage is Fury's
// divided clause: any number of targets from none to four, each chosen target
// getting at least 1 of the 4 (CR 601.2d).
//
// No simplification.
func init() {
	const oracleID = "0961b68d-bb0c-4bfb-ad78-01f90b3b97c8"

	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Amethyst Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Explosive Crystal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(0, 4).Dividing(effects.Divide(4)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
