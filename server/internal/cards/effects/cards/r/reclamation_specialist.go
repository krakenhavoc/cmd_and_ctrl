package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reclamation Specialist — Artifact Creature — Borg {2}{B}, 2/2 (Star Trek
// Commander, #4974):
//
//	"{1}, Sacrifice an artifact creature: Draw a card."
//
// The sacrifice is part of the cost (CR 602.2b, 118.3): any artifact
// creature its controller controls, the Specialist included, and with
// none the ability can't be activated. No tap, so it can be activated any
// number of times at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c348d24-af26-4171-bfa7-44ce9bd36f06",
		Name:         "Reclamation Specialist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact creature: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeN(1, "an artifact creature", effects.Artifact(), effects.Creature())),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
