package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Frostbeast — Creature — Elemental Beast {3}{G}{W}, 2/4 (Ice Age, S82 OLD, #5636):
//
//	"At end of combat, destroy all creatures blocking or blocked by this creature."
//
// Dread Wight's end-of-combat partners: every combat's end, the creatures blocking the Frostbeast or
// blocked by it, read from the Frostbeast as it is or as it last existed (CR 608.2h), are destroyed.
// Regeneration and indestructible apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39050823-ccda-43c0-818d-d5db15d82b99",
		Name:         "Kjeldoran Frostbeast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEndCombat, "Kjeldoran Frostbeast — destroy all creatures blocking or blocked by it", effects.S82OldB842DestroyCombatPartners),
		},
	})
}
