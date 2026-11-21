package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Preposterous Proportions — Sorcery {5}{G}{G} (Foundations, #4465):
//
//	"Creatures you control get +10/+10 and gain vigilance until end of
//	 turn."
//
// Overrun's shape: the boost and the keyword land on the creatures
// you control as it resolves; one that enters later is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e03aad2-6426-4d18-a0e5-bc3d8e6b7b11",
		Name:         "Preposterous Proportions",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{
				Match:     yours,
				Power:     10,
				Toughness: 10,
				Label:     "Preposterous Proportions — +10/+10",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"vigilance"},
				Label:    "Preposterous Proportions — vigilance",
			}.Apply(ctx)
		},
	})
}
