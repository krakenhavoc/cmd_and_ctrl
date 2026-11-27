package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning Fields — Sorcery {4}{R} (Portal Three Kingdoms, S82 #5622):
//
//	"Burning Fields deals 5 damage to target opponent or planeswalker."
//
// The one target is an opponent or a planeswalker (Burning Sun's Avatar's clause),
// re-judged as the spell resolves (CR 608.2b). The damage is dealt by the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3b58231-d532-423b-91d0-78d9c0f66f0f",
		Name:         "Burning Fields",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, err := effects.OldB708DamageToTargetOpponentOrPlaneswalker(ctx, 5)
			return err
		},
	})
}
