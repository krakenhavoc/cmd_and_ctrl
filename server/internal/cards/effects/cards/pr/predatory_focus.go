package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predatory Focus — Sorcery {3}{G}{G} (Guildpact, S82 OLD batch 573, #5599):
//
//	"You may have creatures you control assign their combat damage this
//	 turn as though they weren't blocked."
//
// Garruk, Savage Herald's −7 shape: the creatures you control are locked as
// the spell resolves (CR 611.2c) and each gets the option of CR 510.1c
// (Characteristic.AssignsAsThoughUnblocked) until end of turn; the combat
// damage step asks each one's controller whether to use it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5866507-ef3b-446a-99ea-d31d974bdb01",
		Name:         "Predatory Focus",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
				"Predatory Focus — assign combat damage as though unblocked", game.AssignsAsThoughUnblockedMod())
		})),
	})
}
