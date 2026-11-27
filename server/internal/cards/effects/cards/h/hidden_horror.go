package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Horror — Creature — Horror {1}{B}{B}, 4/4 (Portal Second Age, S82 older sets batch 734, #5625):
//
//	"When this creature enters, sacrifice it unless you discard a creature card."
//
// With no creature card in hand it is sacrificed at once; otherwise its controller picks a creature card to discard or none, which sacrifices it if it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db7c5a2f-9aea-4b14-86cf-7af79d6484af",
		Name:         "Hidden Horror",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hidden Horror — sacrifice it unless you discard a creature card", effects.S82OldB734SacrificeUnlessDiscardACreature("Hidden Horror")),
		},
	})
}
