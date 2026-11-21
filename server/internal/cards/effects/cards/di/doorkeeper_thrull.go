package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doorkeeper Thrull — {1}{W} Creature — Thrull 1/2:
//
//	"Flash
//	 Flying
//	 Artifacts and creatures entering don't cause abilities to trigger."
//
// Hushwing Gryff with the filter widened to artifacts (#1735). A
// noncreature artifact's own "when this enters" is stopped too, and so
// is a Treasure being created for a "whenever an artifact enters"
// watcher (CR 701.7a puts the token onto the battlefield).
func init() {
	s := effects.SuppressesEntering(effects.Or(effects.Artifact(), effects.Creature()))
	s.Label = "Doorkeeper Thrull"
	effects.Register(effects.Spec{
		OracleID:           "0eaea4fd-a377-4541-9bdd-14921034a975",
		Name:               "Doorkeeper Thrull",
		Completeness:       effects.CompletenessFull,
		PrintedKeywords:    []string{"flash", "flying"},
		TriggerSuppressors: []game.TriggerSuppressor{s},
	})
}
