package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Theorix Metamage // Omit Variables — Creature — Shade Wizard
// {2}{U/B}, 2/3 // Sorcery {U/B} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 Threshold — This creature gets +1/+0 and has flying as long as
//	 there are seven or more cards in your graveyard."
//
//	Omit Variables — "Mill three cards."
//
// Two statics (layer 7c and layer 6) behind the same live threshold read.
//
// No simplification.
func init() {
	const id = "92a0aa44-0ef7-4e6f-8288-7c33067c5fd1"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Theorix Metamage",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Static: []game.StaticAbility{
			effects.FraThresholdSelfPT(1, 0),
			effects.FraThresholdSelfKeywords("flying"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Omit Variables",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OmitVariablesResolve,
	})
}
