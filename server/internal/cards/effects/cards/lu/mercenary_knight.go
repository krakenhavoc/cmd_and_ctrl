package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mercenary Knight — Creature — Human Mercenary Knight {2}{B}, 4/4 (Portal,
// S82 OLD batch 783, #5630):
//
//	"When this creature enters, sacrifice it unless you discard a creature card."
//
// Hidden Horror's tax: with no creature card in hand it is sacrificed at once;
// otherwise its controller picks a creature card to discard, or none, which
// sacrifices it if it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed5429bb-233a-4528-bf7d-df5f6b192b1c",
		Name:         "Mercenary Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mercenary Knight — sacrifice it unless you discard a creature card", effects.S82OldB734SacrificeUnlessDiscardACreature("Mercenary Knight")),
		},
	})
}
