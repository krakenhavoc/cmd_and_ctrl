package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spoils of Adventure — Instant {4}{W}{U} (Zendikar Rising, S82 OLD #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior, and
//	 Wizard.)
//	 You gain 3 life and draw three cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "c4f3a6c6-a123-43ad-9513-d65fbff10d02",
		Name:              "Spoils of Adventure",
		Completeness:      effects.CompletenessFull,
		Purpose:           game.Purpose{Draws: 3},
		SelfCostModifiers: []game.CostModifier{effects.S82OldB089PartyCostReduction()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GainLife{Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 3}.Apply(ctx)
		},
	})
}
