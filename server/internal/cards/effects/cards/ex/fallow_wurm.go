package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fallow Wurm — Creature — Wurm {2}{G}, 4/4 (Weatherlight, S82 older sets batch 777, #5629):
//
//	"When this creature enters, sacrifice it unless you discard a land card."
//
// Hidden Horror's tax over land cards: with no land card in hand it is sacrificed at once;
// otherwise its controller picks a land card to discard or none, which sacrifices it if it is
// still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b414f70-a9b1-42c3-9655-6aa520d3b19b",
		Name:         "Fallow Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fallow Wurm — sacrifice it unless you discard a land card", effects.S82OldB777SacrificeUnlessDiscardALand("Fallow Wurm")),
		},
	})
}
