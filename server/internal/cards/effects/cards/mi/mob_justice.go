package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mob Justice — Sorcery {1}{R} (Stronghold, S82 OLD batch 747, #5627):
//
//	"Mob Justice deals damage to target player or planeswalker equal to the
//	 number of creatures you control."
//
// The creatures are counted as the spell resolves, so one that died in
// response is not counted, and the target is read again then (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ff97869-f27c-4379-92fd-605062aa5418",
		Name:         "Mob Justice",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageToTargetEqualToCreaturesYouControl(ctx.Game, item)
		},
	})
}
