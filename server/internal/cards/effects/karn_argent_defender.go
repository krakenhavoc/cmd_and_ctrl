package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karn, Argent Defender — Legendary Artifact Creature — Golem {2},
// 1/3:
//
//	"Artifacts and creatures entering the battlefield don't cause
//	 abilities to trigger."
//
// Torpor Orb's line widened to artifacts as well as creatures: one
// trigger suppressor over Artifact-or-Creature entering (#1735). It
// stops the entering permanent's own "when this enters" and every
// "whenever a creature or artifact enters", and it covers a token being
// created. A permanent that is both is one entry, suppressed once.
//
// No simplification.
func init() {
	s := SuppressesEntering(Or(Artifact(), Creature()))
	s.Label = "Karn, Argent Defender"
	Register(Spec{
		OracleID:           "aecc621f-67da-41a7-9d47-05ded302ae35",
		Name:               "Karn, Argent Defender",
		Completeness:       CompletenessFull,
		TriggerSuppressors: []game.TriggerSuppressor{s},
	})
}
