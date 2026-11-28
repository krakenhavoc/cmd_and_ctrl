package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Wurm — Creature — Wurm {2}{G}, 4/4 (Portal, S82 OLD batch 783,
// #5630):
//
//	"When this creature enters, sacrifice it unless you discard a land card."
//
// Hidden Horror's tax over land cards: with no land card in hand it is
// sacrificed at once; otherwise its controller picks a land card to discard, or
// none, which sacrifices it if it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdab50a8-1e02-4ba5-8c09-e2837e7652f7",
		Name:         "Thundering Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thundering Wurm — sacrifice it unless you discard a land card", effects.S82OldB783SacrificeUnlessDiscardALand("Thundering Wurm")),
		},
	})
}
