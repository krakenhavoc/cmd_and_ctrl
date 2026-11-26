package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crumbling Colossus — Artifact Creature — Golem {5}, 7/4 (Magic 2012, S82
// OLD batch 452, #5573):
//
//	"Trample
//	 When this creature attacks, sacrifice it at end of combat."
//
// Runaway Carriage's delayed sacrifice at the beginning of this combat's
// end of combat step (CR 511.2), on the attack trigger only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7f72e54-7e3a-4720-be69-fc8c5eb4c190",
		Name:            "Crumbling Colossus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Crumbling Colossus — sacrifice it at end of combat",
				effects.SacrificeThisAtEndOfCombat("Crumbling Colossus — sacrifice it")),
		},
	})
}
