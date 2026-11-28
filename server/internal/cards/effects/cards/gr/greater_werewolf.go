package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greater Werewolf — Creature — Werewolf {4}{B}, 2/4 (Homelands, S82 OLD b822,
// #5635):
//
//	"At end of combat, put a -0/-2 counter on each creature blocking or blocked
//	by this creature."
//
// Dread Wight's shape: a trigger at the beginning of every end of combat step
// (CR 511.2) that reads the Werewolf's combat partners as it resolves, from its
// last-known information if it has left (CR 608.2h). The -0/-2 counter is a
// power/toughness counter (CR 122.1a): the engine reads any "[+-]N/[+-]M" kind.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02847b0d-03a1-4e30-902e-60bdf6952c58",
		Name:         "Greater Werewolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEndCombat,
				"Greater Werewolf — put a -0/-2 counter on each creature blocking or blocked by it",
				effects.OldB822EachCombatPartnerGetsACounter("-0/-2")),
		},
	})
}
