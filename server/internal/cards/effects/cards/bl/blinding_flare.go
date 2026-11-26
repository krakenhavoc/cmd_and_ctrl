package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinding Flare — Sorcery {R} (Journey into Nyx):
//
//	"Strive — This spell costs {R} more to cast for each target beyond the first.
//	 Any number of target creatures can't block this turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c8b0a8a-8f80-4c3e-8f27-fb1df94aa0d4",
		Name:         "Blinding Flare",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{R}", "Strive — This spell costs {R} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.S82OldB366LegalTargetCards(ctx) {
				if err := (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Blinding Flare — can't block"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
