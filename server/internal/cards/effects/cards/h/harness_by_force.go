package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harness by Force — Sorcery {1}{R}{R} (Journey into Nyx):
//
//	"Strive — This spell costs {2}{R} more to cast for each target beyond the first.
//	 Gain control of any number of target creatures until end of turn. Untap
//	 those creatures. They gain haste until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c7e2596-6c21-4ce0-95dc-5b4dccb4090a",
		Name:         "Harness by Force",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{2}{R}", "Strive — This spell costs {2}{R} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.S82OldB366LegalTargetCards(ctx) {
				if err := effects.ThreatenAndGrant(ctx, id, "Harness by Force", "haste"); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
